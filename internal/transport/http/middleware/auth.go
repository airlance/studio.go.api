package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/transport/http/utils"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	configuration := ory.NewConfiguration()
	configuration.Servers = ory.ServerConfigurations{
		{
			URL: cfg.Kratos.PublicURL,
		},
	}
	client := ory.NewAPIClient(configuration)

	return func(c *gin.Context) {
		sessionCookie, err := c.Cookie("ory_kratos_session")
		if err != nil {
			utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "Missing session cookie")
			c.Abort()
			return
		}

		// Use the cookie to get the session from Kratos
		session, _, err := client.FrontendAPI.ToSession(context.Background()).Cookie("ory_kratos_session=" + sessionCookie).Execute()
		if err != nil {
			utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "Invalid session")
			c.Abort()
			return
		}

		if !session.GetActive() {
			utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "Session is inactive")
			c.Abort()
			return
		}

		// Check if the user is verified
		identity := session.GetIdentity()
		if identity.Id == "" {
			utils.RespondError(c, http.StatusUnauthorized, "SNAKE_CASE_UNAUTHORIZED", "No identity associated with session")
			c.Abort()
			return
		}

		verifiableAddresses := identity.VerifiableAddresses
		isVerified := false

		for _, addr := range verifiableAddresses {
			if addr.Verified {
				isVerified = true
				break
			}
		}

		if !isVerified {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "SNAKE_CASE_USER_NOT_VERIFIED",
				"message": "Please verify your email to access the dashboard",
				"user":    &identity,
			})
			c.Abort()
			return
		}

		// Store identity in context for later use
		c.Set("user", &identity)
		c.Next()
	}
}
