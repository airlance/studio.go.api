package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/resoul/studio.go.api/internal/app"
	"github.com/resoul/studio.go.api/internal/httpx"
	"github.com/resoul/studio.go.api/internal/ory"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd, args)
	},
}

func serve(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	container, err := app.NewContainer(ctx)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to build app container")
	}
	defer func() {
		if err := container.Close(); err != nil {
			logrus.WithError(err).Error("Error closing database")
		}
	}()

	router := gin.Default()
	router.Use(httpx.CORSMiddleware(container.Config.Server.GetCORSAllowedOrigins()))

	hydraHandler := ory.NewHydraHandler(container.Config.Auth.HydraAdminURL)
	auth := router.Group("/auth/hydra")
	{
		auth.GET("/login", hydraHandler.GetLoginRequest)
		auth.POST("/login/accept", hydraHandler.AcceptLoginRequest)
		auth.POST("/login/reject", hydraHandler.RejectLoginRequest)
		auth.GET("/consent", hydraHandler.GetConsentRequest)
		auth.POST("/consent/accept", hydraHandler.AcceptConsentRequest)
		auth.POST("/consent/reject", hydraHandler.RejectConsentRequest)
	}

	addr := fmt.Sprintf(":%s", container.Config.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logrus.Infof("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-ctx.Done()

	logrus.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("Server forced to shutdown")
	}

	logrus.Info("Server exited")
}
