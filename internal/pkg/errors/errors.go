package errors

import "errors"

var (
	ErrInvalidRequestPayload = errors.New("invalid request payload")

	ErrInternalServerError = errors.New("internal server error")

	ErrInsufficientStock           = errors.New("insufficient stock for this quantity")
	ErrInvalidUserInput            = errors.New("invalid user input")
	ErrProductNotBelongToSeller    = errors.New("product does not belong to this seller")
	ErrInvalidProductUpdatePayload = errors.New("all required columns must not be empty and valid for update")
	ErrProductOutOfStock           = errors.New("product out of stock")

	ErrCartNotFound          = errors.New("cart item not found")
	ErrInvalidCartOperation  = errors.New("invalid cart operation")
	ErrCartAlreadyCheckedOut = errors.New("cart is already checked out")
	ErrCartRetrievalFail     = errors.New("failed to retrieve cart")
	ErrCartEmpty             = errors.New("cart is empty")

	MsgFailedToClearProductCaches = "failed to clear product cache"
	MsgProductCacheCleared        = "product cache cleared"

	ErrCartItemNotFound = errors.New("cart item not found")

	ErrNotFound = errors.New("not found")

	ErrOrderNotFound = errors.New("order not found")

	ErrInvalidUserSession = errors.New("invalid user session")
	ErrUnauthorized       = errors.New("unauthorized")
)
