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
	"github.com/resoul/studio.go.api/internal/middleware"
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

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			httpx.RespondOK(c, gin.H{"status": "ok"})
		})

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(container.Config))
		{
			protected.GET("/user/me", func(c *gin.Context) {
				user, _ := c.Get("user")
				httpx.RespondOK(c, user)
			})
		}
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
