package repositories

import (
	"context"
	"fmt"

	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/db"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *db.InsertProductParams) (*db.Product, error)
	GetAllProducts(ctx context.Context) ([]db.GetAllProductsRow, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*db.GetProductByIDRow, error)
	GetProductByIDs(ctx context.Context, ids []uuid.UUID) ([]db.GetProductByIDsRow, error)
	GetProductsBySellerID(ctx context.Context, sellerID uuid.UUID) ([]db.GetProductsBySellerIDRow, error)
	GetProductsByName(ctx context.Context, name string) ([]db.GetProductsByNameRow, error)
	UpdateProduct(ctx context.Context, updateParams *db.UpdateProductParams) (*db.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) (*db.Product, error)
}

type productRepository struct {
	db  *db.Queries
	log *logrus.Logger
}

func NewProductRepository(sqlcQueries *db.Queries, log *logrus.Logger) ProductRepository {
	return &productRepository{
		db:  sqlcQueries,
		log: log,
	}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *db.InsertProductParams) (*db.Product, error) {
	row, err := r.db.InsertProduct(ctx, *product)
	if err != nil {
		r.log.WithField("product_id", product.ID).WithError(err).Error("Failed to create product in the database")
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return &row, err
}

func (r *productRepository) GetAllProducts(ctx context.Context) ([]db.GetAllProductsRow, error) {
	var rows []db.GetAllProductsRow

	rows, err := r.db.GetAllProducts(ctx)
	if err != nil {
		r.log.WithFields(logrus.Fields{"error": err}).Error("Failed to receive orders from DB")
		return nil, err
	}

	return rows, nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id uuid.UUID) (*db.GetProductByIDRow, error) {
	var row db.GetProductByIDRow

	row, err := r.db.GetProductByID(ctx, id)
	if err != nil {
		r.log.WithFields(logrus.Fields{"id": id, "error": err}).Error("Failed to receive product from DB")
		return nil, fmt.Errorf("failed to receive product from DB: %w", err)
	}

	return &row, nil
}

func (r *productRepository) GetProductByIDs(ctx context.Context, ids []uuid.UUID) ([]db.GetProductByIDsRow, error) {
	var rows []db.GetProductByIDsRow

	rows, err := r.db.GetProductByIDs(ctx, ids)
	if err != nil {
		r.log.WithFields(logrus.Fields{"product_ids": ids, "error": err}).Error("Failed to receive product from DB")
		return nil, err
	}

	return rows, nil
}

func (r *productRepository) GetProductsBySellerID(ctx context.Context, sellerID uuid.UUID) ([]db.GetProductsBySellerIDRow, error) {
	var rows []db.GetProductsBySellerIDRow

	rows, err := r.db.GetProductsBySellerID(ctx, sellerID)
	if err != nil {
		r.log.WithFields(logrus.Fields{"seller_id": sellerID}).WithError(err).Error("Failed to receive products by seller ID from DB")
		return nil, err
	}

	return rows, nil
}

func (r *productRepository) GetProductsByName(ctx context.Context, name string) ([]db.GetProductsByNameRow, error) {
	var rows []db.GetProductsByNameRow

	searchPattern := "%" + name + "%"

	rows, err := r.db.GetProductsByName(ctx, searchPattern)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"name_query": name,
			"error":      err,
		}).Error("Failed to execute the GetProductsByName query in the database")
		return nil, err
	}

	r.log.WithFields(logrus.Fields{"name": name, "error": err}).Error("Product not found")
	return rows, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, updateParams *db.UpdateProductParams) (*db.Product, error) {
	var row db.Product

	row, err := r.db.UpdateProduct(ctx, *updateParams)

	if err != nil {
		r.log.WithField("product_id", updateParams.ID).WithError(err).Error("Failed to update product in the database")
		return nil, err
	}

	return &row, nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id uuid.UUID) (*db.Product, error) {
	var row db.Product

	row, err := r.db.DeleteProduct(ctx, id)
	if err != nil {
		r.log.WithField("product_id", id).WithError(err).Error("Failed to delete product in the database")
		return nil, err
	}

	return &row, nil
}
