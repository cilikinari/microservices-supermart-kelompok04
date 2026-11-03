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

// 1. Implementasi GetStockItem dengan membaca data aktual dari service/DB
func (h *GRPCHandler) GetStockItem(ctx context.Context, req *inventorypb.GetStockItemRequest) (*inventorypb.GetStockItemResponse, error) {
	stock, err := h.service.GetStockByProductAndWarehouse(req.GetProductId(), req.GetWarehouseId())
	if err != nil || stock == nil {
		// Jika data tidak ditemukan, kembalikan item dengan stok 0 sesuai spesifikasi gRPC
		return &inventorypb.GetStockItemResponse{
			Item: &inventorypb.StockItem{
				ProductId:        req.GetProductId(),
				WarehouseId:      req.GetWarehouseId(),
				QuantityOnHand:   0,
				QuantityReserved: 0,
				UnitPrice: &inventorypb.Money{
					CurrencyCode: "IDR",
					Amount:       0,
				},
			},
		}, nil
	}

	return &inventorypb.GetStockItemResponse{
		Item: &inventorypb.StockItem{
			ProductId:        stock.ProductID,
			WarehouseId:      stock.WarehouseID,
			QuantityOnHand:   int32(stock.QuantityOnHand),
			QuantityReserved: int32(stock.QuantityReserved),
			UnitPrice: &inventorypb.Money{
				CurrencyCode: "IDR",
				Amount:       int64(stock.ProductPrice),
			},
		},
	}, nil
}

// 2. Implementasi BatchGetStockItems dengan pemetaan batch
func (h *GRPCHandler) BatchGetStockItems(ctx context.Context, req *inventorypb.BatchGetStockItemsRequest) (*inventorypb.BatchGetStockItemsResponse, error) {
	var items []*inventorypb.StockItem
	for _, pid := range req.GetProductIds() {
		stock, err := h.service.GetStockByProductAndWarehouse(pid, req.GetWarehouseId())
		if err == nil && stock != nil {
			items = append(items, &inventorypb.StockItem{
				ProductId:        stock.ProductID,
				WarehouseId:      stock.WarehouseID,
				QuantityOnHand:   int32(stock.QuantityOnHand),
				QuantityReserved: int32(stock.QuantityReserved),
				UnitPrice: &inventorypb.Money{
					CurrencyCode: "IDR",
					Amount:       int64(stock.ProductPrice),
				},
			})
		}
	}
	return &inventorypb.BatchGetStockItemsResponse{
		Items: items,
	}, nil
}
