package routes

import (
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/delivery/http/middlewares"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/handlers"

	"github.com/labstack/echo/v4"
)

// The `InitRoutes` function sets up various routes for handling product and cart operations with
// authentication and role-based access control in a Go application using Echo framework.
func InitRoutes(e *echo.Echo, handler *handlers.API, authMiddleware echo.MiddlewareFunc) {
	authGroup := e.Group("/api/v1")
	authGroup.Use(authMiddleware)

	productGroup := authGroup.Group("/products")
	productGroup.GET("/", handler.GetAllProducts())
	productGroup.GET("/tag/:tag", handler.GetProductsByName())
	productGroup.GET("/:id", handler.GetProductByID(), middlewares.RequireRoles("admin"))
	productGroup.GET("/seller/:seller_id", handler.GetProductsBySellerID())
	productGroup.POST("/create", handler.CreateProduct(), middlewares.RequireRoles("admin", "seller"))
	productGroup.PUT("/update/:product_id", handler.UpdateProduct(), middlewares.RequireRoles("admin", "seller"))
	productGroup.DELETE("/delete/:product_id", handler.DeleteProduct(), middlewares.RequireRoles("admin", "seller"))
	productGroup.DELETE("/clear-cache", handler.ClearProductCaches())

	cartGroup := authGroup.Group("/cart")
	cartGroup.GET("/", handler.GetCartItemsByUserID())
	cartGroup.POST("/add/:product_id", handler.AddToCart())
	cartGroup.PUT("/update/:product_id", handler.UpdateCartItem())
	cartGroup.DELETE("/remove/:product_id", handler.RemoveFromCart())
}
