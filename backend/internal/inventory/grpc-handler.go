package inventory

import (
	context "context"

	inventorypb "github.com/nusantara-supermart/backend/internal/inventory/pb/inventory/v1"
)

// GRPCHandler mengimplementasikan interface gRPC dari file pb
type GRPCHandler struct {
	inventorypb.UnimplementedInventoryServiceServer
	service Service // Menggunakan layer 'Service' bisnis yang sama dengan REST handler!
}

func NewGRPCHandler(service Service) *GRPCHandler {
	return &GRPCHandler{service: service}
}

// 1. Implementasi GetStockItem
func (h *GRPCHandler) GetStockItem(ctx context.Context, req *inventorypb.GetStockItemRequest) (*inventorypb.GetStockItemResponse, error) {
	// Panggil logika bisnis yang sudah ada dari h.service
	// Contoh:
	// stocks, err := h.service.GetStockByID(req.GetProductId(), req.GetWarehouseId())

	return &inventorypb.GetStockItemResponse{
		Item: &inventorypb.StockItem{
			ProductId:   req.GetProductId(),
			WarehouseId: req.GetWarehouseId(),
			// Petakan data dari service/DB ke struct protobuf di sini
		},
	}, nil
}

// 2. Implementasi BatchGetStockItems
func (h *GRPCHandler) BatchGetStockItems(ctx context.Context, req *inventorypb.BatchGetStockItemsRequest) (*inventorypb.BatchGetStockItemsResponse, error) {
	// Panggil logika bisnis dari h.service
	return &inventorypb.BatchGetStockItemsResponse{
		// Petakan array item di sini
	}, nil
}
