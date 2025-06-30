package middlewares

import (
	"net/http"

	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/models"

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

// import (
// 	"log"
// 	"net/http"
// 	"strings"

// 	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/authclient" // Impor AuthClient
// 	"github.com/labstack/echo/v4"
// )

// // AuthMiddlewareOptions berisi dependensi untuk middleware autentikasi
// type AuthMiddlewareOptions struct {
// 	AuthClient *authclient.AuthClient
// }

// func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		authHeader := c.Request().Header.Get("Authorization")
// 		if authHeader == "" {
// 			return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Token otorisasi tidak ditemukan"})
// 		}

// 		// Asumsi token adalah Bearer <token>
// 		token := authHeader
// 		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
// 			token = authHeader[7:]
// 		} else {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Format token tidak valid (harap gunakan Bearer token)"})
// 		}

// 		isValid, userID, username, role, errMsg, err := a.AuthGRPCClient.ValidateToken(token)
// 		if err != nil {
// 			log.Printf("Kesalahan saat validasi token gRPC: %v", err)
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Kesalahan server saat validasi token"})
// 		}

// 		if !isValid {
// 			return c.JSON(http.StatusUnauthorized, echo.Map{"message": errMsg})
// 		}

// 		c.Set("user_id", userID)
// 		c.Set("username", username)
// 		c.Set("role", role)
// 		log.Printf("Pengguna %s (ID: %s, Peran: %s) berhasil diautentikasi.", username, userID, role)

// 		return next(c)
// 	}
// }

// AuthMiddleware adalah fungsi middleware untuk memvalidasi token
// func AuthMiddleware(opts AuthMiddlewareOptions) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			authHeader := c.Request().Header.Get("Authorization")
// 			if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
// 				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token otentikasi tidak ditemukan atau format salah"})
// 			}
// 			token := authHeader[7:]

// 			// Signature fungsi diubah untuk mendapatkan 'userRole'
// 			isValid, userID, username, userRole, errMsg, err := opts.AuthClient.ValidateToken(token)
// 			if err != nil {
// 				c.Logger().Errorf("Kesalahan validasi token gRPC: %v", err)

// 				// Cek error dari gRPC
// 				st, ok := status.FromError(err)
// 				if ok {
// 					if st.Code() == codes.Unauthenticated {
// 						// Ini akan menangani pesan error seperti "Token kedaluwarsa" atau "Token tidak valid" dari server gRPC
// 						return c.JSON(http.StatusUnauthorized, map[string]string{"message": st.Message()})
// 					}
// 				}
// 				// Generic internal server error for other unexpected errors
// 				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Kesalahan server saat memvalidasi token: " + err.Error()})
// 			}

// 			if !isValid {
// 				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token tidak valid: " + errMsg})
// 			}

// 			// Jika token valid, Anda bisa menyimpan informasi pengguna di Echo Context
// 			c.Set("user_id", userID)
// 			c.Set("username", username)
// 			c.Set("role", userRole)                                                                   // Set 'role' di context
// 			c.Logger().Debugf("User %s (ID: %s, Role: %s) authenticated", username, userID, userRole) // Opsional: log info user

// 			return next(c)
// 		}
// 	}
// }
