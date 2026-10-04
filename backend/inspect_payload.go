package main

import (
	"encoding/json"
	"fmt"
	"log"

	inventorypb "github.com/nusantara-supermart/backend/internal/inventory/pb/inventory/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	// 1. Objek Representasi gRPC (Protobuf)
	respProto := &inventorypb.GetStockItemResponse{
		Item: &inventorypb.StockItem{
			ProductId:        "p0000001-0000-0000-0000-000000000001",
			WarehouseId:      "w0000001-0000-0000-0000-000000000001",
			QuantityOnHand:   100,
			QuantityReserved: 5,
		},
	}

	protoBytes, err := proto.Marshal(respProto)
	if err != nil {
		log.Fatalf("Gagal serialize protobuf: %v", err)
	}

	// 2. Objek Representasi REST (JSON)
	respJSON := map[string]interface{}{
		"item": map[string]interface{}{
			"product_id":        "p0000001-0000-0000-0000-000000000001",
			"warehouse_id":      "w0000001-0000-0000-0000-000000000001",
			"quantity_on_hand":  100,
			"quantity_reserved": 5,
		},
	}
	
	// jsonBytes murni untuk dihitung ukurannya (tanpa spasi/enter agar akurat seperti di jaringan)
	jsonBytes, err := json.Marshal(respJSON)
	if err != nil {
		log.Fatalf("Gagal serialize JSON: %v", err)
	}

	// prettyJSON untuk dicetak rapi ke laporan
	prettyJSON, _ := json.MarshalIndent(respJSON, "", "  ")

	// 3. Cetak Hasil Komparasi
	fmt.Println("=============== CONTOH PAYLOAD JSON AKTUAL ===============")
	fmt.Println(string(prettyJSON))
	
	fmt.Println("\n=============== ANALISIS WIRE-SIZE PAYLOAD ===============")
	fmt.Printf("Ukuran Payload JSON (HTTP/1.1) : %d Bytes\n", len(jsonBytes))
	fmt.Printf("Ukuran Payload gRPC (Protobuf) : %d Bytes\n", len(protoBytes))
	
	selisih := len(jsonBytes) - len(protoBytes)
	persentase := (float64(selisih) / float64(len(jsonBytes))) * 100
	fmt.Printf("Selisih Penghematan Bandwidth  : %d Bytes (Lebih ringkas %.2f%%)\n", selisih, persentase)
}