package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/entities"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/helpers"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/models"
	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/pkg/errors"
)

func (api *API) CreateProduct() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		api.log.Infof("Received request to create product from IP: %s", c.RealIP())

		userID, err := getUserIDFromContext(c)
		if err != nil {
			return respondError(c, http.StatusUnauthorized, errors.ErrInvalidUserSession)
		}

		var req models.ProductRequest
		if err := c.Bind(&req); err != nil {
			return respondError(c, http.StatusBadRequest, errors.ErrInvalidRequestPayload)
		}

		res, err := api.ProductSvc.CreateProduct(ctx, userID, &req)
		if err != nil {
			return handleOperationError(c, err, MsgFailedToCreateProduct)
		}

		return respondSuccess(c, http.StatusCreated, MsgProductCreated, toProductResponse(res))
	}
}

func (api *API) GetAllProducts() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		c.Logger().Infof("Received request for product list from IP: %s", c.RealIP())

		res, err := api.ProductSvc.GetAllProducts(ctx)
		if err != nil {
			return handleGetError(c, err)
		}

		return respondSuccess(c, http.StatusOK, MsgProductRetrieved, toProductResponseList(res))
	}
}

func (api *API) GetProductsByName() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		c.Logger().Infof("Received request for GetProductsByName from IP: %s", c.RealIP())

		ProductName, err := getFromPathParam(c, "tag")
		if err != nil {
			return respondError(c, http.StatusBadRequest, err)
		}

		res, err := api.ProductSvc.GetProductsByName(ctx, ProductName)
		if err != nil {
			return handleGetError(c, err)
		}

		return respondSuccess(c, http.StatusOK, MsgProductRetrieved, toProductResponseList(res))
	}
}

func (api *API) GetProductByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		c.Logger().Infof("Received request for GetProductByID from IP: %s", c.RealIP())

		productID, err := getIDFromPathParam(c, "id")
		if err != nil {
			return respondError(c, http.StatusBadRequest, err)
		}

		res, err := api.ProductSvc.GetProductByID(ctx, productID)
		if err != nil {
			return handleGetError(c, err)
		}

		return respondSuccess(c, http.StatusOK, MsgProductRetrieved, toProductResponse(res))
	}
}

func (api *API) GetProductsBySellerID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		c.Logger().Infof("Received request for GetProductsBySellerID from IP: %s", c.RealIP())

		sellerID, err := getIDFromPathParam(c, "seller_id")
		if err != nil {
			return respondError(c, http.StatusBadRequest, err)
		}

		res, err := api.ProductSvc.GetProductsBySellerID(ctx, sellerID)
		if err != nil {
			return handleGetError(c, err)
		}

		return respondSuccess(c, http.StatusOK, MsgProductRetrieved, toProductResponseList(res))
	}
}

func (api *API) UpdateProduct() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		c.Logger().Infof("Received request for UpdateProduct from IP: %s", c.RealIP())

		userID, err := getUserIDFromContext(c)
		if err != nil {
			return respondError(c, http.StatusUnauthorized, errors.ErrInvalidUserSession)
		}

		productID, err := getIDFromPathParam(c, "product_id")
		if err != nil {
			return respondError(c, http.StatusBadRequest, err)
		}

		var productData models.ProductRequest
		if err := c.Bind(&productData); err != nil {
			return respondError(c, http.StatusBadRequest, errors.ErrInvalidRequestPayload)
		}

		res, err := api.ProductSvc.UpdateProduct(ctx, &productData, productID, userID)
		if err != nil {
			return handleOperationError(c, err, MsgFailedToUpdateProduct)
		}

		return respondSuccess(c, http.StatusOK, MsgProductUpdated, toProductResponse(res))

	}
}

func (api *API) DeleteProduct() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		c.Logger().Infof("Received request for DeleteProduct from IP: %s", c.RealIP())

		sellerID, err := getUserIDFromContext(c)
		if err != nil {
			return respondError(c, http.StatusUnauthorized, errors.ErrInvalidUserSession)
		}

		productID, err := getIDFromPathParam(c, "product_id")
		if err != nil {
			return respondError(c, http.StatusBadRequest, err)
		}

		res, err := api.ProductSvc.DeleteProduct(ctx, productID, sellerID)
		if err != nil {
			return handleOperationError(c, err, MsgFailedToDeleteProduct)
		}

		return respondSuccess(c, http.StatusOK, MsgProductDeleted, toProductResponse(res))
	}
}

func (api *API) ClearProductCaches() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		c.Logger().Infof("Received request for ClearProductCaches from IP: %s", c.RealIP())

		err := api.ProductSvc.ResetAllProductCaches(ctx)
		if err != nil {
			return handleOperationError(c, err, errors.MsgFailedToClearProductCaches)
		}

		return respondSuccess(c, http.StatusOK, errors.MsgProductCacheCleared, nil)
	}
}

// ------- HELPER -------
func toProductResponse(product *entities.Product) *models.ProductResponse {
	return &models.ProductResponse{
		ID:          product.ID,
		SellerID:    product.SellerID,
		Name:        product.Name,
		Price:       product.Price,
		Stock:       product.Stock,
		Discount:    product.Discount,
		Type:        product.Type,
		Description: product.Description,
		CreatedAt:   product.CreatedAt.Format(helpers.LAYOUTFORMAT),
		UpdatedAt:   product.UpdatedAt.Format(helpers.LAYOUTFORMAT),
	}
}

func toProductResponseList(products []entities.Product) []*models.ProductResponse {
	var productResponses []*models.ProductResponse

	for _, product := range products {
		productResponses = append(productResponses, toProductResponse(&product))
	}

	return productResponses
}
