package catalog

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventorypb "github.com/nusantara-supermart/backend/internal/inventory/pb/inventory/v1"
)

type InventoryClient struct {
	client inventorypb.InventoryServiceClient
}

func NewInventoryClient(grpcAddress string) *InventoryClient {
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: Gagal connect ke gRPC Inventory: %v", err)
		return nil
	}
	client := inventorypb.NewInventoryServiceClient(conn)
	return &InventoryClient{client: client}
}

// GetStockInfo memanggil RPC GetStockItem untuk mengambil data stok produk real-time dari Inventory (upstream)
func (c *InventoryClient) GetStockInfo(ctx context.Context, productID, warehouseID string) (int32, int32, error) {
	if c == nil || c.client == nil {
		return 0, 0, nil
	}
	req := &inventorypb.GetStockItemRequest{
		ProductId:   productID,
		WarehouseId: warehouseID,
	}

	res, err := c.client.GetStockItem(ctx, req)
	if err != nil {
		return 0, 0, err
	}

	item := res.GetItem()
	if item == nil {
		return 0, 0, nil
	}

	return item.GetQuantityOnHand(), item.GetQuantityReserved(), nil
}

