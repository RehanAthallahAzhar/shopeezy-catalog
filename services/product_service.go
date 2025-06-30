package services

import (
	"context"
	stdErrors "errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/helpers"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/models"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/errors"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/repositories"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]models.ProductResponse, error)
	GetProductsBySellerID(ctx context.Context, sellerID string) ([]models.ProductResponse, error)
	GetProductByID(ctx context.Context, id string) (*models.ProductResponse, error)
	GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error)
	CreateProduct(ctx context.Context, userID string, productData *models.ProductRequest) ([]models.ProductResponse, error)
	UpdateProduct(ctx context.Context, productID string, productData *models.ProductRequest, sellerID string) ([]models.ProductResponse, error)
	DeleteProduct(ctx context.Context, productID string, sellerID string) error
}

type productServiceImpl struct {
	productRepo repositories.ProductRepository
	validator   *validator.Validate
}

// Dependency Injection
func NewProductService(productRepo repositories.ProductRepository, validator *validator.Validate) ProductService {
	return &productServiceImpl{
		productRepo: productRepo,
		validator:   validator,
	}
}

func (s *productServiceImpl) GetAllProducts(ctx context.Context) ([]models.ProductResponse, error) {
	products, err := s.productRepo.GetAllProducts(ctx)
	if err != nil {
		// spesific error
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("service: failed to retrieve all products: %w", err)
	}
	if len(products) == 0 {
		return nil, errors.ErrProductNotFound
	}
	var res []models.ProductResponse
	for _, product := range products {
		res = append(res, *mapToProductResponse(&product))
	}

	return res, nil
}

func (s *productServiceImpl) GetProductsBySellerID(ctx context.Context, sellerID string) ([]models.ProductResponse, error) {
	products, err := s.productRepo.GetProductsBySellerID(ctx, sellerID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("service: failed to retrieve products by seller ID %s: %w", sellerID, err)
	}

	if len(products) == 0 {
		return nil, errors.ErrProductNotFound
	}

	var res []models.ProductResponse
	for _, product := range products {
		res = append(res, *mapToProductResponse(&product))
	}

	return res, nil
}

func (s *productServiceImpl) GetProductsByName(ctx context.Context, name string) ([]models.ProductResponse, error) {
	products, err := s.productRepo.GetProductsByName(ctx, name)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("service: failed to retrieve products by seller ID %s: %w", name, err)
	}

	if len(products) == 0 {
		return nil, errors.ErrProductNotFound
	}

	var res []models.ProductResponse
	for _, product := range products {
		res = append(res, *mapToProductResponse(&product))
	}

	return res, nil
}

func (s *productServiceImpl) GetProductByID(ctx context.Context, id string) (*models.ProductResponse, error) {
	product, err := s.productRepo.GetProductByID(ctx, id)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return nil, err
		}

		return nil, fmt.Errorf("service: failed to retrieve all products: %w", err)
	}

	return mapToProductResponse(product), nil
}

func (s *productServiceImpl) CreateProduct(ctx context.Context, userID string, productData *models.ProductRequest) ([]models.ProductResponse, error) {
	if err := s.validator.Struct(productData); err != nil {
		validationErrors := err.(validator.ValidationErrors)

		var errorMessages []string
		for _, fieldErr := range validationErrors {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed on the '%s' tag", fieldErr.Field(), fieldErr.Tag()))
		}

		log.Printf("Product validation failed: %v", errorMessages)
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidRequestPayload, strings.Join(errorMessages, ", ")) // Menggunakan error kustom
	}

	product := &models.Product{
		ID:          helpers.GenerateNewID(),
		SellerID:    userID,
		Name:        productData.Name,
		Price:       productData.Price,
		Stock:       productData.Stock,
		Discount:    productData.Discount,
		Type:        productData.Type,
		Description: productData.Description,
	}

	err := s.productRepo.CreateProduct(ctx, product)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return nil, errors.ErrProductNotFound
		}

		return nil, fmt.Errorf("service: failed to add product: %w", err)
	}

	createdProduct, err := s.GetAllProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: failed to retrieve created product: %w", err)
	}

	return createdProduct, nil
}

func (s *productServiceImpl) UpdateProduct(ctx context.Context, productID string, productData *models.ProductRequest, sellerID string) ([]models.ProductResponse, error) {
	if err := s.validator.Struct(productData); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		var errorMessages []string
		for _, fieldErr := range validationErrors {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed on the '%s' tag", fieldErr.Field(), fieldErr.Tag()))
		}
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidRequestPayload, strings.Join(errorMessages, ", "))
	}

	// Check if the product exists and belongs to the same seller id
	existingProduct, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return nil, errors.ErrProductNotFound
		}

		return nil, fmt.Errorf("service: failed to find product for update: %w", err)
	}

	if existingProduct.SellerID != sellerID {
		return nil, fmt.Errorf("service: product does not belong to this seller")
	}

	product := &models.Product{
		ID:       productID,
		SellerID: sellerID,
		Name:     productData.Name,
		Price:    productData.Price,
		Stock:    productData.Stock,
		Discount: productData.Discount,
	}

	err = s.productRepo.UpdateProduct(ctx, product)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return nil, errors.ErrProductNotFound
		}

		return nil, fmt.Errorf("service: failed to update product: %w", err)
	}

	updatedProduct, err := s.GetAllProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: failed to retrieve created product: %w", err)
	}

	return updatedProduct, nil
}

func (s *productServiceImpl) DeleteProduct(ctx context.Context, productID string, sellerID string) error {
	// Check if the product exists and belongs to the same seller
	existingProduct, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return errors.ErrProductNotFound
		}

		return fmt.Errorf("service: failed to find product for deletion: %w", err)
	}

	if existingProduct.SellerID != sellerID {
		return fmt.Errorf("service: product does not belong to this seller")
	}

	err = s.productRepo.DeleteProduct(ctx, productID)
	if err != nil {
		if stdErrors.Is(err, errors.ErrProductNotFound) {
			return errors.ErrProductNotFound
		}

		return fmt.Errorf("service: failed to delete product: %w", err)
	}

	return nil
}

func mapToProductResponse(product *models.ProductWithSeller) *models.ProductResponse {
	return &models.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Price:       product.Price,
		Stock:       product.Stock,
		Discount:    product.Discount,
		Type:        product.Type,
		Description: product.Description,
		SellerID:    product.SellerID,
		SellerName:  product.SellerName,
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}
}
