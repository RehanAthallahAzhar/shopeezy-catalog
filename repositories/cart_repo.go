package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/models"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/errors"
	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/redisclient"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type contextKey string

const userIDKey contextKey = "userID"

type CartRepository interface {
	GetCartItemsByUserID(ctx context.Context, userID string) ([]models.CartItem, error)
	GetCartItemByProductID(ctx context.Context, userID, productID string) (*models.CartItem, error)
	// GetCartItemsByUsername(ctx context.Context, Username string) ([]models.CartItem, error) //coming soon
	AddItemToCart(ctx context.Context, userID string, cartID string, cartData *models.Cart) error
	UpdateItemQuantity(ctx context.Context, productID string, userID string, req *models.Cart) error
	RemoveItemFromCart(ctx context.Context, userID string, productID string) error
	RestoreCartFromDB(ctx context.Context, userID string) error
	ClearCart(ctx context.Context, userID string) error
	CheckoutCart(ctx context.Context, userID string) error
}

type cartRepository struct {
	db          *gorm.DB
	productRepo ProductRepository
	redisClient *redisclient.RedisClient
}

func NewCartRepository(db *gorm.DB, productRepo ProductRepository, redisClient *redisclient.RedisClient) CartRepository {
	return &cartRepository{db: db, productRepo: productRepo, redisClient: redisClient}
}

func (r *cartRepository) getCartKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
}

func (r *cartRepository) GetCartItemsByUserID(ctx context.Context, userID string) ([]models.CartItem, error) {
	/*
		Get list of cart on redis
	*/

	cartKey := r.getCartKey(userID)

	// Retrieve all fields and values from Redis Hash
	cartMap, err := r.redisClient.Client.HGetAll(ctx, cartKey).Result()
	if err != nil {
		return nil, errors.ErrCartRetrievalFail
	}

	var cartItems []models.CartItem
	for productID, jsonStr := range cartMap {
		var item models.RedisCartItem
		err := json.Unmarshal([]byte(jsonStr), &item)
		if err != nil {
			log.Printf("Warning: Failed to parse cart item %s: %v", productID, err)
			continue
		}

		// Retrieve product details from the (cached) ProductRepository
		product, err := r.productRepo.GetProductByID(ctx, productID)
		if err != nil {
			log.Printf("Warning: Failed to get product details %s for user's cart %s: %v", productID, userID, err)
			continue // Proceed to the next item if the product is not found
		}

		cartItems = append(cartItems, models.CartItem{
			ID:              cartKey,
			SellerID:        product.SellerID,
			SellerName:      product.Name,
			Quantity:        item.Quantity,
			ProductID:       productID,
			ProductName:     product.Name,
			ProductPrice:    product.Price,
			CartDescription: item.CartDescription,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		})
	}
	return cartItems, nil
}

func (r *cartRepository) GetCartItemByProductID(ctx context.Context, userID, productID string) (*models.CartItem, error) {
	cartKey := r.getCartKey(userID)

	// Ambil nilai dari Redis hash
	jsonStr, err := r.redisClient.Client.HGet(ctx, cartKey, productID).Result()
	if err == redis.Nil {
		return nil, errors.ErrCartItemNotFound // custom error: "cart item not found"
	} else if err != nil {
		return nil, fmt.Errorf("failed to get cart item from Redis: %w", err)
	}

	var redisItem models.RedisCartItem
	if err := json.Unmarshal([]byte(jsonStr), &redisItem); err != nil {
		return nil, fmt.Errorf("failed to parse cart item JSON: %w", err)
	}

	// Ambil detail produk dari repo produk
	product, err := r.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to retrieve product info: %w", err)
	}

	return &models.CartItem{
		ID:              cartKey,
		SellerID:        product.SellerID,
		SellerName:      product.Name,
		Quantity:        redisItem.Quantity,
		ProductID:       productID,
		ProductName:     product.Name,
		ProductPrice:    product.Price,
		CartDescription: redisItem.CartDescription,
		CreatedAt:       redisItem.CreatedAt,
		UpdatedAt:       redisItem.UpdatedAt,
	}, nil
}

func (r *cartRepository) AddItemToCart(ctx context.Context, userID string, cartID string, cartData *models.Cart) error {
	cartKey := r.getCartKey(userID)

	currentStr, err := r.redisClient.Client.HGet(ctx, cartKey, cartData.ProductID).Result()
	var currentQuantity int
	var existingItem *models.RedisCartItem

	if err == nil {
		var parsedItem models.RedisCartItem
		if err := json.Unmarshal([]byte(currentStr), &parsedItem); err == nil {
			currentQuantity = parsedItem.Quantity
			existingItem = &parsedItem
		}
	} else if err != redis.Nil {
		return fmt.Errorf("failed to retrieve an item from the Redis carts: %w", err)
	}

	newQuantity := currentQuantity + cartData.Quantity
	if newQuantity <= 0 {
		return r.RemoveItemFromCart(ctx, userID, cartData.ProductID)
	}

	cartRedis := NewRedisCartItem(newQuantity, cartData.Description, existingItem)

	cartRedisJSON, err := json.Marshal(cartRedis)
	if err != nil {
		return fmt.Errorf("failed to marshal cart item: %w", err)
	}

	// Store to Redis
	pipe := r.redisClient.Client.Pipeline()
	pipe.HSet(ctx, cartKey, cartData.ProductID, cartRedisJSON)
	pipe.Expire(ctx, cartKey, 24*time.Hour)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add/update item to Redis cart: %w", err)
	}
	// Async backup ke database
	go func() {
		bgCtx := context.WithValue(context.Background(), userIDKey, userID)
		err := r.backupCartItemToDB(bgCtx, cartID, userID, cartData.ProductID, newQuantity, cartData.Description)
		if err != nil {
			log.Printf("failed to backup cart item to DB (user: %s, product: %s): %v", userID, cartData.ProductID, err)
		}
	}()

	log.Printf("User %s added product %s (qty: %d) to cart. Total: %d", userID, cartData.ProductID, cartData.Quantity, newQuantity)
	return nil
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, productID, userID string, req *models.Cart) error {
	if req.Quantity <= 0 {
		return errors.ErrInvalidRequestPayload
	}

	// Validate whether the product exists
	product, err := r.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCartNotFound
		}
		return fmt.Errorf("product not found: %w", err)
	}

	// Enough stock validation
	if product.Stock < req.Quantity {
		return errors.ErrInvalidRequestPayload
	}

	cartKey := r.getCartKey(userID)

	// Retrieve old items from Redis to maintain CreatedAt
	var existingItem *models.RedisCartItem
	currentStr, err := r.redisClient.Client.HGet(ctx, cartKey, productID).Result()
	if err == nil {
		var parsedItem models.RedisCartItem
		if unmarshalErr := json.Unmarshal([]byte(currentStr), &parsedItem); unmarshalErr == nil {
			existingItem = &parsedItem
		}
	} else if err != redis.Nil {
		return fmt.Errorf("failed to retrieve existing cart item: %w", err)
	}

	// Create a new item with a new quantity and description
	newItem := NewRedisCartItem(req.Quantity, req.Description, existingItem)
	itemJSON, err := json.Marshal(newItem)
	if err != nil {
		return fmt.Errorf("failed to marshal updated cart item: %w", err)
	}

	// Save to Redis
	pipe := r.redisClient.Client.Pipeline()
	pipe.HSet(ctx, cartKey, productID, itemJSON)
	pipe.Expire(ctx, cartKey, 24*time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to update Redis cart item: %w", err)
	}

	// Async backup to DB
	go func() {
		bgCtx := context.WithValue(context.Background(), userIDKey, userID)
		err := r.backupCartItemToDB(bgCtx, productID, userID, productID, req.Quantity, req.Description)
		if err != nil {
			log.Printf("failed to backup cart item to DB (user: %s, product: %s): %v", userID, productID, err)
		}
	}()

	log.Printf("User %s updated product %s to qty: %d", userID, productID, req.Quantity)
	return nil
}

func (r *cartRepository) RemoveItemFromCart(ctx context.Context, userID string, productID string) error {
	cartKey := r.getCartKey(userID)

	if err := r.redisClient.Client.HDel(ctx, cartKey, productID).Err(); err != nil {
		return fmt.Errorf("failed to remove item from Redis cart: %w", err)
	}

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&models.Cart{}).Error; err != nil {
		log.Printf("warning: failed to delete cart item from DB for user %s and product %s: %v", userID, productID, err)
	}

	log.Printf("User %s removed product %s from cart", userID, productID)
	return nil
}

func (r *cartRepository) ClearCart(ctx context.Context, userID string) error {
	cartKey := r.getCartKey(userID)

	// Delete the entire cart in Redis
	if err := r.redisClient.Client.Del(ctx, cartKey).Err(); err != nil {
		return fmt.Errorf("failed to clear cart from Redis: %w", err)
	}

	// Delete all cart entries from user's DB
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.Cart{}).Error; err != nil {
		log.Printf("warning: failed to clear cart from DB for user %s: %v", userID, err)
	}

	log.Printf("User %s cleared the entire cart", userID)
	return nil
}

func (r *cartRepository) CheckoutCart(ctx context.Context, userID string) error {
	// Get all items from Redis cart
	cartItems, err := r.GetCartItemsByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrCartNotFound
		}

		return fmt.Errorf("failed to get cart items for checkout: %w", err)
	}
	if len(cartItems) == 0 {
		return errors.ErrCartNotFound
	}

	// Start a database transaction (if you want to save the order to the main DB)
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin database transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create a new Order in the database
	order := &models.Order{
		UserID:    userID, // Adjust to your user ID type
		Status:    "Pending",
		OrderDate: time.Now(),
		// TotalAmount will be calculated from items
	}
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create new order: %w", err)
	}

	var totalAmount int
	// Add OrderItems to the database and decrement stock
	for _, item := range cartItems {
		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.ProductPrice, // Price at checkout time
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create order item: %w", err)
		}

		// Decrement stock in the database
		if err := r.productRepo.DecrementProductStock(ctx, item.ProductID, item.Quantity); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to decrement stock for product %s: %w", item.ProductName, err)
		}

		if err := r.RemoveItemFromCart(ctx, userID, item.ProductID); err != nil {
			log.Printf("Warning: Failed to clear user %s's cart from Redis after checkout: %v", userID, err)
		}

		totalAmount += item.ProductPrice * item.Quantity
	}

	order.TotalAmount = totalAmount
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order total: %w", err)
	}

	// Commit the database transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("User %s successfully checked out order ID: %s", userID, order.ID)
	return nil
}

func (r *cartRepository) backupCartItemToDB(ctx context.Context, cartID string, userID, productID string, quantity int, desc string) error {
	var existing models.Cart
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		First(&existing).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(&models.Cart{
			ID:          cartID,
			ProductID:   productID,
			UserID:      userID,
			Quantity:    quantity,
			Description: desc,
		}).Error
	}

	existing.Quantity = quantity
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *cartRepository) RestoreCartFromDB(ctx context.Context, userID string) error {
	var items []models.Cart
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	if err != nil {
		return err
	}

	cartKey := r.getCartKey(userID)
	pipe := r.redisClient.Client.Pipeline()
	for _, item := range items {
		pipe.HSet(ctx, cartKey, item.ProductID, item.Quantity)
	}
	pipe.Expire(ctx, cartKey, 24*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func NewRedisCartItem(quantity int, description string, existing *models.RedisCartItem) *models.RedisCartItem {
	now := time.Now()

	item := &models.RedisCartItem{
		Quantity:        quantity,
		CartDescription: description,
		UpdatedAt:       now,
	}

	if existing != nil {
		item.CreatedAt = existing.CreatedAt
	} else {
		item.CreatedAt = now
	}

	return item
}
