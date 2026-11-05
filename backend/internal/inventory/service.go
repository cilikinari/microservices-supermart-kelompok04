package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	SetCatalogClient(client *CatalogClient)
	GetStocks(warehouseID string) ([]InventoryStock, error)
	GetStockByProductAndWarehouse(productID, warehouseID string) (*InventoryStock, error)
	AdjustStock(stockID string, req AdjustStockRequest) error
	ReserveStock(productID, warehouseID, orderID string, qty int) error
	ReleaseStock(orderID string) error
	CreateMutation(req StockMutationRequest) error
	GetMutations() ([]StockMutation, error)
	GetWarehouses() ([]Warehouse, error)
	CreateWarehouse(req WarehouseRequest) (*Warehouse, error)
	GetWarehouseZones(warehouseID string) ([]WarehouseZone, error)
	GetLowStockAlerts() ([]LowStockAlert, error)
	GetOpnames(warehouseID string) ([]StockOpname, error)
	CreateOpname(userID string, req CreateOpnameRequest) (*StockOpname, error)
}

type inventoryService struct {
	repo          Repository
	catalogClient *CatalogClient
}

func NewService(repo Repository, catalogClient ...*CatalogClient) Service {
	s := &inventoryService{repo: repo}
	if len(catalogClient) > 0 && catalogClient[0] != nil {
		s.catalogClient = catalogClient[0]
	}
	return s
}

func (s *inventoryService) SetCatalogClient(client *CatalogClient) {
	s.catalogClient = client
}

func (s *inventoryService) GetStocks(warehouseID string) ([]InventoryStock, error) {
	stocks, err := s.repo.GetStocks(warehouseID)
	if err != nil {
		return nil, err
	}

	// Enrichment via gRPC Catalog Client (Catalog Upstream/Supplier -> Inventory Downstream/Customer, pola Conformist)
	if s.catalogClient != nil && len(stocks) > 0 {
		productIDs := make([]string, 0, len(stocks))
		for _, st := range stocks {
			if st.ProductID != "" {
				productIDs = append(productIDs, st.ProductID)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		prodMap, err := s.catalogClient.BatchGetProducts(ctx, productIDs)
		if err == nil {
			for i := range stocks {
				if prod, ok := prodMap[stocks[i].ProductID]; ok && prod != nil {
					stocks[i].ProductTitle = prod.Title
					stocks[i].ProductSKU = prod.Sku
					stocks[i].ProductPrice = prod.BasePrice
				}
			}
		}
	}

	return stocks, nil
}

func (s *inventoryService) GetStockByProductAndWarehouse(productID, warehouseID string) (*InventoryStock, error) {
	stock, err := s.repo.GetStockByProductAndWarehouse(productID, warehouseID)
	if err != nil {
		return nil, err
	}

	// Enrichment via gRPC Catalog Client
	if s.catalogClient != nil && stock != nil && stock.ProductID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		prod, err := s.catalogClient.GetProduct(ctx, stock.ProductID)
		if err == nil && prod != nil {
			stock.ProductTitle = prod.Title
			stock.ProductSKU = prod.Sku
			stock.ProductPrice = prod.BasePrice
		}
	}

	return stock, nil
}

func (s *inventoryService) AdjustStock(stockID string, req AdjustStockRequest) error {
	if req.QuantityOnHand < 0 {
		return errors.New("quantity on hand cannot be negative")
	}
	return s.repo.UpdateStockQuantity(stockID, req.QuantityOnHand)
}

func (s *inventoryService) ReserveStock(productID, warehouseID, orderID string, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}
	return s.repo.ReserveStock(productID, warehouseID, orderID, qty)
}

func (s *inventoryService) ReleaseStock(orderID string) error {
	return s.repo.ReleaseStock(orderID)
}

func (s *inventoryService) CreateMutation(req StockMutationRequest) error {
	if req.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	if req.SourceWarehouseID == req.DestinationWarehouseID {
		return errors.New("source and destination warehouse must be different")
	}

	mutation := &StockMutation{
		ProductID:              req.ProductID,
		SourceWarehouseID:      req.SourceWarehouseID,
		DestinationWarehouseID: req.DestinationWarehouseID,
		Quantity:               req.Quantity,
	}

	return s.repo.CreateStockMutation(mutation)
}

func (s *inventoryService) GetMutations() ([]StockMutation, error) {
	mutations, err := s.repo.GetStockMutations()
	if err != nil {
		return nil, err
	}

	// Enrichment via gRPC Catalog Client
	if s.catalogClient != nil && len(mutations) > 0 {
		productIDs := make([]string, 0, len(mutations))
		for _, m := range mutations {
			if m.ProductID != "" {
				productIDs = append(productIDs, m.ProductID)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		prodMap, err := s.catalogClient.BatchGetProducts(ctx, productIDs)
		if err == nil {
			for i := range mutations {
				if prod, ok := prodMap[mutations[i].ProductID]; ok && prod != nil {
					mutations[i].ProductTitle = prod.Title
				}
			}
		}
	}

	return mutations, nil
}

func (s *inventoryService) GetWarehouses() ([]Warehouse, error) {
	return s.repo.GetWarehouses()
}

func (s *inventoryService) CreateWarehouse(req WarehouseRequest) (*Warehouse, error) {
	if req.WarehouseCode == "" || req.WarehouseName == "" {
		return nil, errors.New("warehouse_code and warehouse_name are required")
	}
	w := &Warehouse{
		ID:            uuid.NewString(),
		WarehouseCode: req.WarehouseCode,
		WarehouseName: req.WarehouseName,
		Address:       req.Address,
		City:          req.City,
	}
	if err := s.repo.CreateWarehouse(w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *inventoryService) GetWarehouseZones(warehouseID string) ([]WarehouseZone, error) {
	return s.repo.GetWarehouseZones(warehouseID)
}

func (s *inventoryService) GetLowStockAlerts() ([]LowStockAlert, error) {
	alerts, err := s.repo.GetLowStockAlerts()
	if err != nil {
		return nil, err
	}

	// Enrichment via gRPC Catalog Client
	if s.catalogClient != nil && len(alerts) > 0 {
		productIDs := make([]string, 0, len(alerts))
		for _, a := range alerts {
			if a.ProductID != "" {
				productIDs = append(productIDs, a.ProductID)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		prodMap, err := s.catalogClient.BatchGetProducts(ctx, productIDs)
		if err == nil {
			for i := range alerts {
				if prod, ok := prodMap[alerts[i].ProductID]; ok && prod != nil {
					alerts[i].ProductTitle = prod.Title
				}
			}
		}
	}

	return alerts, nil
}

func (s *inventoryService) GetOpnames(warehouseID string) ([]StockOpname, error) {
	return s.repo.GetStockOpnames(warehouseID)
}

func (s *inventoryService) CreateOpname(userID string, req CreateOpnameRequest) (*StockOpname, error) {
	opNumber := fmt.Sprintf("OPN-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
	opDate := req.OpnameDate
	if opDate == "" {
		opDate = time.Now().Format("2006-01-02")
	}

	op := &StockOpname{
		ID:                uuid.NewString(),
		WarehouseID:       req.WarehouseID,
		OpnameNumber:      opNumber,
		OpnameDate:        opDate,
		ConductedByUserID: userID,
	}

	if err := s.repo.CreateStockOpname(op); err != nil {
		return nil, err
	}
	return op, nil
}
