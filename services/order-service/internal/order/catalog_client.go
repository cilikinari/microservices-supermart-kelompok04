package order

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	catalogpb "github.com/nusantara-supermart/proto/catalog/v1"
)

// CatalogClient menangani komunikasi gRPC sinkron dari Order Service ke Catalog Service
type CatalogClient struct {
	client catalogpb.CatalogServiceClient
	conn   *grpc.ClientConn
}

func NewCatalogClient(grpcAddress string) *CatalogClient {
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("WARNING: Order Service gagal connect ke gRPC Catalog di %s: %v", grpcAddress, err)
		return nil
	}

	return &CatalogClient{
		client: catalogpb.NewCatalogServiceClient(conn),
		conn:   conn,
	}
}

func (c *CatalogClient) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *CatalogClient) GetProduct(ctx context.Context, productID string) (*catalogpb.ProductItem, error) {
	if c == nil || c.client == nil || productID == "" {
		return nil, nil
	}

	res, err := c.client.GetProduct(ctx, &catalogpb.GetProductRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, err
	}
	return res.GetProduct(), nil
}
