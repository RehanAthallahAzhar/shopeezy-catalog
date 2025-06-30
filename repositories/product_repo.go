package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/models"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/redisclient"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ProductRepository interface {
	GetAllProducts(ctx context.Context) ([]models.ProductWithSeller, error)
	GetProductByID(ctx context.Context, id string) (*models.ProductWithSeller, error)
	GetProductsBySellerID(ctx context.Context, id string) ([]models.ProductWithSeller, error)
	GetProductsByName(ctx context.Context, name string) ([]models.ProductWithSeller, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
	DecrementProductStock(ctx context.Context, productID string, quantity int) error
	IncrementProductStock(ctx context.Context, productID string, quantity int) error
	InvalidateProductCache(ctx context.Context, productID string)
}

type productRepository struct {
	db          *gorm.DB
	redisClient *redisclient.RedisClient
}

func NewProductRepository(db *gorm.DB, redisClient *redisclient.RedisClient) ProductRepository {
	return &productRepository{db: db, redisClient: redisClient}
}

func (r *productRepository) GetAllProducts(ctx context.Context) ([]models.ProductWithSeller, error) {
	allProductsKey := "all_products"
	var products []models.ProductWithSeller

	val, err := r.redisClient.Client.Get(ctx, allProductsKey).Result()
	if err == nil {
		err = json.Unmarshal([]byte(val), &products)
		if err == nil {
			log.Println("All products are retrieved from Redis cache")
			return products, nil
		}
		log.Printf("Error unmarshalling all products from Redis: %v, will fetch from DB", err)
	} else if err != redis.Nil {
		log.Printf("Error getting all products from Redis: %v, will fetch from DB", err)
	}

	err = r.db.WithContext(ctx).Table("products").
		Select(`products.id, products.seller_id, users.name as seller_name, products.name, products.price,
	        products.stock, products.discount, products.type, products.description, 
	        products.created_at, products.updated_at`).
		Joins("LEFT JOIN users ON users.id = products.seller_id").
		Where("products.deleted_at is null").Find(&products).Error
	if err != nil {
		return nil, err
	}

	productsJSON, err := json.Marshal(products)
	if err != nil {
		log.Printf("Failed to  marshal product for Redis cache: %v", err)
	} else {
		err = r.redisClient.Client.Set(ctx, allProductsKey, productsJSON, 2*time.Minute).Err() // Cache duration
		if err != nil {
			log.Printf("Failed to store product to Redis cache: %v", err)
		} else {
			log.Println("All products have been saved to Redis cache")
		}
	}

	return products, nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id string) (*models.ProductWithSeller, error) {
	productKey := fmt.Sprintf("product:%s", id)
	var product models.ProductWithSeller

	// Try to get from Redis cache
	val, err := r.redisClient.Client.Get(ctx, productKey).Result()
	if err == nil {
		// Cache hit!
		err = json.Unmarshal([]byte(val), &product)
		if err == nil {
			log.Printf("Product %s retrieved from Redis cache.", id)
			return &product, nil
		}
		log.Printf("Error unmarshalling product from Redis: %v, will fetch from DB", err)
	} else if err != redis.Nil {
		log.Printf("Error getting product from Redis: %v, will fetch from DB", err)
	}

	// fetch from database
	result := r.db.WithContext(ctx).Table("products").
		Select(`products.id, products.seller_id, users.name as seller_name, products.name, products.price,
	        products.stock, products.discount, products.type, products.description, 
	        products.created_at, products.updated_at`).
		Joins("LEFT JOIN users ON users.id = products.seller_id").
		Where("products.id = ? AND products.deleted_at is null", id).First(&product)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("produk tidak ditemukan")
		}
		return nil, result.Error
	}

	// Save to Redis cache for future requests
	productJSON, err := json.Marshal(product)
	if err != nil {
		log.Printf("Failed to  marshal product %s for Redis cache: %v", id, err)
	} else {
		// Cache for 5 minutes
		err = r.redisClient.Client.Set(ctx, productKey, productJSON, 5*time.Minute).Err() // 5 minutes
		if err != nil {
			log.Printf("Failed to store product %s to Redis cache: %v", id, err)
		} else {
			log.Printf("product %s successfully saved in redis cache.", id)
		}
	}

	return &product, nil
}

func (r *productRepository) GetProductsBySellerID(ctx context.Context, sellerID string) ([]models.ProductWithSeller, error) {
	sellerProductKey := fmt.Sprintf("products_by_seller:%s", sellerID)
	var products []models.ProductWithSeller

	val, err := r.redisClient.Client.Get(ctx, sellerProductKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(val), &products); err == nil {
			log.Printf("Products by seller %s retrieved from Redis cache", sellerID)
			return products, nil
		}
		log.Printf("Error unmarshalling seller products from Redis: %v", err)
	} else if err != redis.Nil {
		log.Printf("Error accessing Redis for seller products: %v", err)
	}

	result := r.db.WithContext(ctx).Table("products").
		Select(`products.id, products.seller_id, users.name as seller_name, products.name, products.price,
	        products.stock, products.discount, products.type, products.description, 
	        products.created_at, products.updated_at`).
		Joins("LEFT JOIN users ON users.id = products.seller_id").
		Where("products.seller_id = ? AND products.deleted_at is null", sellerID).First(&products)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("produk tidak ditemukan")
		}
		return nil, result.Error
	}

	productsJSON, err := json.Marshal(products)
	if err != nil {
		log.Printf("Failed to marshal products for seller %s: %v", sellerID, err)
	} else {
		err = r.redisClient.Client.Set(ctx, sellerProductKey, productsJSON, 5*time.Minute).Err()
		if err != nil {
			log.Printf("Failed to cache products for seller %s: %v", sellerID, err)
		}
	}

	return products, nil
}

func (r *productRepository) GetProductsByName(ctx context.Context, name string) ([]models.ProductWithSeller, error) {
	sellerProductKey := fmt.Sprintf("products_by_name:%s", name)
	var products []models.ProductWithSeller

	val, err := r.redisClient.Client.Get(ctx, sellerProductKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(val), &products); err == nil {
			log.Printf("Products by seller %s retrieved from Redis cache", name)
			return products, nil
		}
		log.Printf("Error unmarshalling seller products from Redis: %v", err)
	} else if err != redis.Nil {
		log.Printf("Error accessing Redis for seller products: %v", err)
	}

	queryName := "%" + name + "%"

	result := r.db.WithContext(ctx).Table("products").
		Select(`products.id, products.seller_id, users.name as seller_name, products.name, products.price,
	        products.stock, products.discount, products.type, products.description, 
	        products.created_at, products.updated_at`).
		Joins("LEFT JOIN users ON users.id = products.seller_id").
		Where("products.name like ? AND products.deleted_at is null", queryName).First(&products)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("produk tidak ditemukan")
		}
		return nil, result.Error
	}

	productsJSON, err := json.Marshal(products)
	if err != nil {
		log.Printf("Failed to marshal products for seller %s: %v", name, err)
	} else {
		err = r.redisClient.Client.Set(ctx, sellerProductKey, productsJSON, 5*time.Minute).Err()
		if err != nil {
			log.Printf("Failed to cache products for seller %s: %v", name, err)
		}
	}

	return products, nil
}

func (r *productRepository) InvalidateProductCache(ctx context.Context, productID string) {
	productKey := fmt.Sprintf("product:%s", productID)
	allProductsKey := "all_products"

	// Clear specific product cache
	err := r.redisClient.Client.Del(ctx, productKey).Err()
	if err != nil {
		log.Printf("Failed to clear the product cache %s: %v", productID, err)
	} else {
		log.Printf("Product cache %s cleared.", productID)
	}

	// Clear cache of all products (cause the list has changed)
	err = r.redisClient.Client.Del(ctx, allProductsKey).Err()
	if err != nil {
		log.Printf("Failed to clear the all products cache: %v", err)
	} else {
		log.Println("Cache of all products cleared.")
	}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	if err := r.db.WithContext(ctx).Table("products").Create(&product).Error; err != nil {
		return err
	}

	r.InvalidateProductCache(ctx, product.ID)

	return nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	if err := r.db.WithContext(ctx).
		Table("products").
		Where("id = ? AND seller_id = ?", product.ID, product.SellerID).
		Updates(&product).Error; err != nil {
		return err
	}

	r.InvalidateProductCache(ctx, product.ID)

	return nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Table("products").Where("id = ? ", id).Delete(&models.Product{}).Error; err != nil {
		return err
	}

	r.InvalidateProductCache(ctx, id)

	return nil
}

func (r *productRepository) DecrementProductStock(ctx context.Context, productID string, quantity int) error {
	var product models.Product
	res := r.db.WithContext(ctx).First(&product, productID)

	if res.Error != nil {
		return res.Error
	}

	if product.Stock < quantity {
		return fmt.Errorf("stok tidak mencukupi")
	}

	product.Stock -= quantity

	return r.UpdateProduct(ctx, &product)
}

func (r *productRepository) IncrementProductStock(ctx context.Context, productID string, quantity int) error {
	var product models.Product
	res := r.db.WithContext(ctx).First(&product, productID)
	if res.Error != nil {
		return res.Error
	}
	product.Stock += quantity
	return r.UpdateProduct(ctx, &product)
}
