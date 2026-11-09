package catalog

import (
	"context"
	"log"

	catalogpb "github.com/nusantara-supermart/backend/internal/catalog/pb/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler mengimplementasikan interface gRPC dari CatalogService
type GRPCHandler struct {
	catalogpb.UnimplementedCatalogServiceServer
	service Service
}

func NewGRPCHandler(service Service) *GRPCHandler {
	return &GRPCHandler{service: service}
}

func mapToPBProduct(p *Product) *catalogpb.ProductItem {
	if p == nil {
		return nil
	}
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}
	categoryName := ""
	if p.CategoryName != nil {
		categoryName = *p.CategoryName
	}
	brandName := ""
	if p.BrandName != nil {
		brandName = *p.BrandName
	}

	return &catalogpb.ProductItem{
		Id:           p.ID,
		Sku:          p.SKU,
		Title:        p.Title,
		Description:  desc,
		BasePrice:    p.BasePrice,
		WeightGram:   int32(p.WeightGram),
		IsPublished:  p.IsPublished,
		CategoryName: categoryName,
		BrandName:    brandName,
	}
}

// GetProduct mengambil detail produk berdasarkan product_id
func (h *GRPCHandler) GetProduct(ctx context.Context, req *catalogpb.GetProductRequest) (*catalogpb.GetProductResponse, error) {
	if req.GetProductId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "product_id is required")
	}

	product, _, _, err := h.service.GetProductByID(req.GetProductId())
	if err != nil || product == nil {
		return nil, status.Errorf(codes.NotFound, "product %s not found", req.GetProductId())
	}

	return &catalogpb.GetProductResponse{
		Product: mapToPBProduct(product),
	}, nil
}

// BatchGetProducts mengambil daftar produk berdasarkan kumpulan product_ids
func (h *GRPCHandler) BatchGetProducts(ctx context.Context, req *catalogpb.BatchGetProductsRequest) (*catalogpb.BatchGetProductsResponse, error) {
	var items []*catalogpb.ProductItem

	for _, pid := range req.GetProductIds() {
		if pid == "" {
			continue
		}
		product, _, _, err := h.service.GetProductByID(pid)
		if err == nil && product != nil {
			items = append(items, mapToPBProduct(product))
		}
	}

	log.Printf("📡 [gRPC Server] Catalog melayani BatchGetProducts: %d produk dikirim ke caller (Inventory)", len(items))
	return &catalogpb.BatchGetProductsResponse{
		Products: items,
	}, nil
}

