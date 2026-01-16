package transport

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"product/api/server/productinternal"
	appservice "product/pkg/product/app/service"
)

func NewProductInternalAPI(
	productService appservice.ProductService,
) productinternal.ProductInternalServiceServer {
	return &productInternalAPI{
		productService: productService,
	}
}

type productInternalAPI struct {
	productService appservice.ProductService

	productinternal.UnimplementedProductInternalServiceServer
}

func (p *productInternalAPI) CreateProduct(ctx context.Context, request *productinternal.CreateProductRequest) (*productinternal.CreateProductResponse, error) {
	if request.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "product name cannot be empty")
	}
	if request.Price < 0 {
		return nil, status.Error(codes.InvalidArgument, "product price cannot be negative")
	}
	if request.Stock < 0 {
		return nil, status.Error(codes.InvalidArgument, "product stock cannot be negative")
	}

	productID, err := p.productService.CreateProduct(ctx, request.Name, request.Price, int(request.Stock))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	return &productinternal.CreateProductResponse{
		ProductID: productID.String(),
	}, nil
}
