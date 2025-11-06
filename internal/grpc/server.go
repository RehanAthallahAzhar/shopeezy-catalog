package grpc

import (
	"context"

	// Import package pb dari repo protos Anda
	product "github.com/RehanAthallahAzhar/shopeezy-protos/pb/product"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Import repository produk Anda

	"github.com/RehanAthallahAzhar/shopeezy-catalog/internal/repositories"
)

// Server struct akan mengimplementasikan interface gRPC Server.
// Ia membutuhkan akses ke repository untuk mengambil data dari database.
type ProductServer struct {
	product.UnimplementedProductServiceServer // Wajib untuk forward compatibility
	ProductRepo                               repositories.ProductRepository
}

// NewProductServer adalah constructor untuk server.
func NewProductServer(productRepo repositories.ProductRepository) *ProductServer {
	return &ProductServer{
		ProductRepo: productRepo,
	}
}

func (s *ProductServer) GetProducts(ctx context.Context, req *product.GetProductsRequest) (*product.GetProductsResponse, error) {
	ids := make([]uuid.UUID, 0, len(req.GetIds()))

	for _, idStr := range req.GetIds() {
		parsedID, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Format Product ID '%s' tidak valid", idStr)
		}
		ids = append(ids, parsedID)
	}

	dbProducts, err := s.ProductRepo.GetProductByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	var protoProducts []*product.Product
	for _, p := range dbProducts {
		protoProducts = append(protoProducts, &product.Product{
			Id:        p.ID.String(),
			SellerId:  p.SellerID.String(),
			Name:      p.Name,
			Price:     int32(p.Price),
			Stock:     int32(p.Stock),
			CreatedAt: timestamppb.New(p.CreatedAt),
			UpdatedAt: timestamppb.New(p.UpdatedAt),
		})
	}

	return &product.GetProductsResponse{Products: protoProducts}, nil
}
