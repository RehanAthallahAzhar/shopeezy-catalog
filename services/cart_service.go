package services

import (
	"context"
	stdErrors "errors"
	"fmt"
	"log"

	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/helpers"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/models"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/errors"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/repositories"
)

type CartService interface {
	GetCartItemsByUserID(ctx context.Context, userID string) ([]models.CartItemResponse, error)
	GetCartItemByProductID(ctx context.Context, userID string, productID string) (*models.CartItemResponse, error)
	AddItemToCart(ctx context.Context, userID string, cartData *models.CartRequest) ([]models.CartItemResponse, error)
	UpdateItemQuantity(ctx context.Context, productID string, userID string, req *models.UpdateCartRequest) ([]models.CartItemResponse, error)
	RemoveItemFromCart(ctx context.Context, userID string, productID string) error
	RestoreCartFromDB(ctx context.Context, userID string) ([]models.CartItemResponse, error)
	ClearCart(ctx context.Context, userID string) error
	CheckoutCart(ctx context.Context, userID string) error
}

type cartServiceImpl struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository) CartService {
	return &cartServiceImpl{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartServiceImpl) GetCartItemsByUserID(ctx context.Context, userID string) ([]models.CartItemResponse, error) {
	cartItems, err := s.cartRepo.GetCartItemsByUserID(ctx, userID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrCartNotFound) {
			return nil, errors.ErrCartNotFound
		}
		log.Printf("Error in service getting cart for user %s: %v", userID, err)
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, errors.ErrProductNotFound
	}

	var res []models.CartItemResponse
	for _, cartItem := range cartItems {
		res = append(res, *mapToCartResponse(userID, &cartItem))
	}

	return res, nil
}

func (s *cartServiceImpl) GetCartItemByProductID(ctx context.Context, userID string, productID string) (*models.CartItemResponse, error) {
	cartItems, err := s.cartRepo.GetCartItemByProductID(ctx, userID, productID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrCartNotFound) {
			return nil, errors.ErrCartNotFound
		}
		log.Printf("Error in service getting cart for user %s: %v", productID, err)
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	return mapToCartResponse(userID, cartItems), nil
}

func (s *cartServiceImpl) AddItemToCart(ctx context.Context, userID string, req *models.CartRequest) ([]models.CartItemResponse, error) {

	product, err := s.productRepo.GetProductByID(ctx, req.ProductID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrCartNotFound) {
			return nil, errors.ErrCartNotFound
		}
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// checking stock
	if product.Stock < req.Quantity {
		return nil, errors.ErrInvalidRequestPayload
	}

	cartID := helpers.GenerateNewID()

	cartData := &models.Cart{
		ID:          cartID,
		ProductID:   req.ProductID,
		UserID:      userID,
		Quantity:    req.Quantity,
		Description: req.Description,
	}

	err = s.cartRepo.AddItemToCart(ctx, userID, cartID, cartData)
	if err != nil {
		log.Printf("Error in service adding item %s to cart for user %s: %v", cartData.ProductID, userID, err)
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}

	createdData, err := s.GetCartItemsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error in service getting created cart for user %s: %v", userID, err)
	}

	return createdData, nil
}

// RemoveItemFromCart menghapus item dari keranjang.
func (s *cartServiceImpl) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	err := s.cartRepo.RemoveItemFromCart(ctx, userID, productID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrCartNotFound) {
			return errors.ErrCartNotFound
		}

		log.Printf("Error in service removing item %s from cart for user %s: %v", productID, userID, err)
		return fmt.Errorf("gagal menghapus item dari keranjang: %w", err)
	}
	return nil
}

func (s *cartServiceImpl) UpdateItemQuantity(ctx context.Context, productID string, userID string, req *models.UpdateCartRequest) ([]models.CartItemResponse, error) {
	cartData := &models.Cart{
		Quantity: req.Quantity,
	}

	if req.Description != "" {
		cartData.Description = req.Description
	}

	err := s.cartRepo.UpdateItemQuantity(ctx, productID, userID, cartData)
	if err != nil {
		if stdErrors.Is(err, errors.ErrCartNotFound) {
			return nil, errors.ErrCartNotFound
		}

		log.Printf("Error in service updating item %s quantity for user %s: %v", productID, userID, err)
		return nil, fmt.Errorf("gagal mengupdate kuantitas item: %w", err)
	}

	updatedCart, err := s.GetCartItemsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error in service getting updated cart for user %s: %v", userID, err)
	}
	return updatedCart, nil
}

func (s *cartServiceImpl) RestoreCartFromDB(ctx context.Context, userID string) ([]models.CartItemResponse, error) {
	err := s.cartRepo.RestoreCartFromDB(ctx, userID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrCartNotFound) {
			return nil, errors.ErrCartNotFound
		}

		log.Printf("Error in service restoring cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to restore cart: %w", err)
	}

	restoredCart, err := s.GetCartItemsByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error in service getting restored cart for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get restored cart: %w", err)
	}

	return restoredCart, nil

}

func (s *cartServiceImpl) ClearCart(ctx context.Context, userID string) error {
	err := s.cartRepo.ClearCart(ctx, userID)
	if err != nil {
		log.Printf("Error in service clearing cart for user %s: %v", userID, err)
		return fmt.Errorf("failed to clean the cart: %w", err)
	}
	return nil
}

func (s *cartServiceImpl) CheckoutCart(ctx context.Context, userID string) error {
	err := s.cartRepo.CheckoutCart(ctx, userID)
	if err != nil {
		log.Printf("Error in service processing checkout for user %s: %v", userID, err)
		return fmt.Errorf("gagal memproses checkout: %w", err)
	}
	return nil
}

func mapToCartResponse(userID string, cart *models.CartItem) *models.CartItemResponse {
	return &models.CartItemResponse{
		ID:          cart.ID,
		ProductID:   cart.ProductID,
		ProductName: cart.ProductName,
		Price:       cart.ProductPrice,
		UserID:      userID,
		Quantity:    cart.Quantity,
		Description: cart.CartDescription,
		CreatedAt:   cart.CreatedAt,
		UpdatedAt:   cart.UpdatedAt,
	}
}
