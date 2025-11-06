package middlewares

import (
	"net/http"
	"strings"

	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/models"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/pkg/grpc/account"
	"github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

func RequireRoles(allowedRoles ...string) echo.MiddlewareFunc {
	roleSet := make(map[string]struct{})
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(string)
			if !ok {
				return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Unauthorized"})
			}

			if _, allowed := roleSet[role]; !allowed {
				return c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "Access denied"})
			}

			return next(c)
		}
	}
}
func AuthMiddleware(authClient *account.AuthClient, log *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Authorization token not found"})
			}

			token := authHeader
			log.Println("Extracted token:", token)
			if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
				token = authHeader[7:]
			} else {
				return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid token format (expected Bearer token)"})
			}

			isValid, userID, username, role, errMsg, err := authClient.ValidateToken(token)
			if err != nil {
				log.Printf("Error during gRPC token validation: %v", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Server error during token validation"})
			}

			if !isValid {
				return c.JSON(http.StatusUnauthorized, echo.Map{"message": errMsg})
			}

			c.Set("userID", userID)
			c.Set("username", username)
			c.Set("role", role)
			log.Printf("User %s (ID: %s, Role: %s) successfully authenticated.", username, userID, role)

			return next(c)
		}
	}
}
