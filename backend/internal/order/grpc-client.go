package order

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Import pb inventory
	inventorypb "github.com/nusantara-supermart/backend/internal/inventory/pb/inventory/v1"
)

type InventoryClient struct {
	client inventorypb.InventoryServiceClient
}

func NewInventoryClient(grpcAddress string) *InventoryClient {
	// 1. Dial / Buat koneksi ke gRPC Server milik Inventory (port :50051)
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Gagal connect ke gRPC Inventory: %v", err)
	}

	// 2. Buat instance client dari pb
	client := inventorypb.NewInventoryServiceClient(conn)
	return &InventoryClient{client: client}
}

// Contoh method Order memanggil gRPC Inventory
func (c *InventoryClient) CheckStock(ctx context.Context, productID string) (int32, error) {
	req := &inventorypb.GetStockItemRequest{
		ProductId: productID,
	}

	// Panggil RPC GetStockItem
	res, err := c.client.GetStockItem(ctx, req)
	if err != nil {
		return 0, err
	}

	return res.GetItem().GetQuantityOnHand(), nil
}