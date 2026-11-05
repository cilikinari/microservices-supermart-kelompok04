package inventory

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	catalogpb "github.com/nusantara-supermart/backend/internal/catalog/pb/catalog/v1"
)

// CatalogClient menangani panggilan gRPC dari Inventory (Downstream / Customer) ke Catalog (Upstream / Supplier)
// Sesuai DDD Context Mapping: Catalog bertindak sebagai Upstream penyedia master data produk.
type CatalogClient struct {
	client catalogpb.CatalogServiceClient
	conn   *grpc.ClientConn
}

func NewCatalogClient(grpcAddress string) *CatalogClient {
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("WARNING: Gagal connect ke gRPC Catalog di %s: %v", grpcAddress, err)
		return nil
	}

	client := catalogpb.NewCatalogServiceClient(conn)
	return &CatalogClient{
		client: client,
		conn:   conn,
	}
}

func (c *CatalogClient) Close() {
	if c != nil && c.conn != nil {
		_ = c.conn.Close()
	}
}

// GetProduct mengambil satu produk via gRPC Catalog
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

// BatchGetProducts mengambil daftar produk sekaligus via gRPC Catalog
func (c *CatalogClient) BatchGetProducts(ctx context.Context, productIDs []string) (map[string]*catalogpb.ProductItem, error) {
	result := make(map[string]*catalogpb.ProductItem)
	if c == nil || c.client == nil || len(productIDs) == 0 {
		return result, nil
	}

	res, err := c.client.BatchGetProducts(ctx, &catalogpb.BatchGetProductsRequest{
		ProductIds: productIDs,
	})
	if err != nil {
		return result, err
	}

	for _, p := range res.GetProducts() {
		if p != nil && p.GetId() != "" {
			result[p.GetId()] = p
		}
	}
	return result, nil
}

