package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/resoul/studio.go.api/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	identityID string
	adminRole  string
)

var adminCmd = cobra.Command{
	Use:   "admin",
	Short: "Admin management commands",
}

var grantAdminCmd = cobra.Command{
	Use:   "grant",
	Short: "Grant admin role to a Kratos identity",
	Run: func(cmd *cobra.Command, args []string) {
		grantAdmin(cmd)
	},
}

type jsonPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	grantAdminCmd.Flags().StringVar(&identityID, "identity-id", "", "Kratos identity ID")
	grantAdminCmd.Flags().StringVar(&adminRole, "role", "admin", "Role value to set in traits.role")
	grantAdminCmd.MarkFlagRequired("identity-id")

	adminCmd.AddCommand(&grantAdminCmd)
}

func grantAdmin(cmd *cobra.Command) {
	ctx := cmd.Context()
	cfg := config.Init(ctx)

	kratosAdminURL := strings.TrimRight(cfg.Auth.KratosAdminURL, "/")
	identityPath := fmt.Sprintf("%s/admin/identities/%s", kratosAdminURL, identityID)

	client := &http.Client{Timeout: 15 * time.Second}

	patch := []jsonPatchOp{
		{
			Op:    "add",
			Path:  "/traits/role",
			Value: adminRole,
		},
	}

	payload, err := json.Marshal(patch)
	if err != nil {
		logrus.WithError(err).Fatal("failed to build JSON patch payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, identityPath, bytes.NewReader(payload))
	if err != nil {
		logrus.WithError(err).Fatal("failed to build request")
	}

	req.Header.Set("Content-Type", "application/json-patch+json")
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Fatal("failed to call Kratos admin API")
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		logrus.Fatalf("kratos admin patch failed with status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	logrus.Infof("Identity %s updated. traits.role=%s", identityID, adminRole)
}
