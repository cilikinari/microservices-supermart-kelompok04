# 🛒 PT Nusantara SuperMart Indonesia - Microservices Platform
> **Tugas Besar Arsitektur Sistem Terdistribusi & Mikroservis**  
> **Kelompok 04**

---

### Rincian Layanan Mikro (*Microservices*) & Basis Data Terisolasi

| Service | DBMS / Broker | Port Host (HTTP / gRPC) | Port DB | Nama Database | Username |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Frontend & Reverse Proxy** | Nginx | `4200` / `80` (HTTP Ingress) | - | - | - |
| **Identity Service** | MySQL 8.0 | `3001` (HTTP) | `3309` | `identity_service_db` | `identity_admin` |
| **Catalog Service** | MySQL 8.0 | `3002` (HTTP), `50052` (gRPC) | `3307` | `catalog_service_db` | `catalog_admin` |
| **Inventory Service**| MySQL 8.0 | `3003` (HTTP), `50051` (gRPC) | `3308` | `inventory_service_db` | `inventory_admin` |
| **Order Service** | MySQL 8.0 | `3004` (HTTP) | `3310` | `order_service_db` | `order_admin` |
| **RabbitMQ Broker** | 3.13 Mgmt | `5672` (AMQP), `15672` (UI) | - | Virtual Host: `/` | `supermart` |

---

## 🚀 Panduan Menjalankan Sistem (Step-by-Step)

### 1. Prasyarat (*Prerequisites*)
Pastikan perangkat kamu sudah terpasang:
* **Docker Desktop** (versi 24+) & **Docker Compose v2**
* **Go** (versi 1.22 atau lebih baru - untuk CLI testing)
* **Git**

### 2. Kloning Repositori & Setup Environment
```bash
# 1. Masuk ke direktori proyek
cd microservices-supermart-kelompok04

# 2. Siapkan file .env (jika belum ada)
cp .env.example .env
```

### 3. Menyalakan Seluruh Kontainer (Full Stack)
Jalankan satu perintah berikut dari direktori root:

```bash
docker compose -f deployments/docker/docker-compose.yml up --build -d
```

### 4. Verifikasi Kesehatan 10 Kontainer
Periksa status kontainer:
```bash
docker compose -f deployments/docker/docker-compose.yml ps
```
Semua 10 kontainer berikut **wajib berstatus `Up` / `healthy`**:
1. `supermart-identity-db`
2. `supermart-catalog-db`
3. `supermart-inventory-db`
4. `supermart-order-db`
5. `supermart-rabbitmq`
6. `supermart-identity-service`
7. `supermart-catalog-service`
8. `supermart-inventory-service`
9. `supermart-order-service`
10. `supermart-frontend`

### 5. Memantau Log Layanan
Untuk melihat proses inisialisasi gRPC, database, atau RabbitMQ consumer per-service:
```bash
# Pantau log Catalog gRPC Server:
docker compose -f deployments/docker/docker-compose.yml logs -f catalog-service

# Pantau log Inventory gRPC Server & RabbitMQ Consumer:
docker compose -f deployments/docker/docker-compose.yml logs -f inventory-service

# Pantau log Order Service & RabbitMQ Publisher:
docker compose -f deployments/docker/docker-compose.yml logs -f order-service
```
---

## 🧪 Skenario Pengujian Lengkap

### Skenario A: Uji Liveness & Healthcheck (REST API via Gateway)
Buka di peramban atau jalankan lewat terminal:
```bash
curl http://localhost:3000/api/v1/health
```
**Ekspektasi Output:**
```json
{
  "success": true,
  "message": "Nusantara SuperMart Microservices Gateway is operational",
  "data": {
    "database": "healthy",
    "services": {
      "catalog": "healthy",
      "identity": "healthy",
      "inventory": "healthy",
      "order": "healthy"
    },
    "version": "1.0.0"
  }
}
```

---

### Skenario B: Uji Event-Driven RabbitMQ
Masuk ke antarmuka RabbitMQ Web Management:
* **URL**: [http://localhost:15672](http://localhost:15672)
* **Username**: `supermart`
* **Password**: `supermart123`
* Di menu **Exchanges**, amati `supermart.events` dan `supermart.dlx`.
* Di menu **Queues**, amati antrean `inventory_stock_reservation_queue` dan `inventory_stock_reservation_queue.dlq`.

Gunakan tool pengujian otomatis yang telah disediakan di `tools/publish-test`:

#### 1. Uji Pengiriman Massal (Batch Mode)
Menguji pengiriman 5 antrean transaksi secara asinkron:
```bash
go run ./tools/publish-test -mode=batch -count=5
```
* **Hasil di Log Inventory Service:** Consumer menerima 5 event pesanan, memotong stok secara aman, dan mengirim sinyal `Ack`.

#### 2. Uji Idempotensi (Mitigasi Double-Reservation)
Menguji apa yang terjadi jika jaringan mengirimkan pesan pesanan yang sama 2 kali:
```bash
go run ./tools/publish-test -mode=duplicate -event-id=DEMO-DUP-001
```
* **Hasil di Log Inventory Service:**  
  * Pesan 1: Diproses dan stok dikunci (*quantity_reserved + 2*).  
  * Pesan 2: Terdeteksi duplikat! Consumer mencetak log:  
    `[Idempotency] Event DEMO-DUP-001 ... sudah pernah diproses. Melewati mutasi stok ganda.`  

#### 3. Uji Poison Pill & Dead-Letter Queue (DLQ)
Menguji ketahanan sistem jika menerima pesan korup atau JSON cacat kontrak:
```bash
go run ./tools/publish-test -mode=poison
```
* **Hasil:** Consumer otomatis menolak pesan via `NACK (requeue=false)` dan RabbitMQ mengalihkannya ke **DLQ** (`inventory_stock_reservation_queue.dlq`) tanpa membuat worker service crash! Cek antrean DLQ di browser [http://localhost:15672/#/queues](http://localhost:15672/#/queues).

---

### Skenario C: Verifikasi Komunikasi gRPC (Bagian Astrid)
Layanan gRPC aktif secara mandiri pada kontainernya masing-masing:
* **Inventory gRPC**: Port `50051` (`supermart-inventory-service`)
* **Catalog gRPC**: Port `50052` (`supermart-catalog-service`)

Saat Frontend membuka etalase katalog atau saat k6 dijalankan, komunikasi gRPC langsung terjadi dan tercetak di log kontainer terkait:
```bash
docker compose -f deployments/docker/docker-compose.yml logs -f catalog-service
docker compose -f deployments/docker/docker-compose.yml logs -f inventory-service
```

---

### Skenario D: Uji Antarmuka Web Frontend
Buka browser di:
* **URL Utama**: [http://localhost:4200](http://localhost:4200) atau [http://localhost:80](http://localhost:80)

1. **Jelajah Katalog**: Buka halaman produk. Frontend mengambil data dari gateway via endpoint `/api/v1/catalog/products`.
2. **Login Akun Demo**:
   * Email: `wayan.putra@gmail.com`
   * Password: `password123`
3. **Simulasi Checkout**:
   * Masukkan produk ke keranjang belanja.
   * Lakukan checkout pesanan.
   * Amati terminal: Event `OrderCreated` otomatis terbit ke RabbitMQ oleh `order-service` dan stok gudang otomatis ter-reservasi di database Inventory oleh `inventory-service`.

---

### Skenario E: Uji Beban & Benchmark Performa k6 (REST vs gRPC)
Jalankan pengujian performa menggunakan skrip k6 yang sudah ada di direktori `scripts/` (50 VUs selama 30 detik):

```bash
# 1. Benchmark REST API (Inventory Stocks)
k6 run scripts/rest_test.js

# 2. Benchmark gRPC Inventory (GetStockItem - Port 50051)
k6 run scripts/grpc_test.js

# 3. Benchmark gRPC Catalog (GetProduct - Port 50052)
k6 run scripts/catalog_grpc_test.js
```

---

## 🛠️ Perintah Berguna (*Cheat Sheet*)

```bash
# 1. Menyalakan semua kontainer:
docker compose -f deployments/docker/docker-compose.yml up --build -d

# 2. Menghentikan semua kontainer:
docker compose -f deployments/docker/docker-compose.yml down

# 3. Menghentikan kontainer dan MERESET SEMUA DATA DATABASE (Fresh Start):
docker compose -f deployments/docker/docker-compose.yml down -v

# 4. Menjalankan mikroservis secara mandiri untuk debugging lokal:
go run ./services/identity-service/cmd/main.go
go run ./services/catalog-service/cmd/main.go
go run ./services/inventory-service/cmd/main.go
go run ./services/order-service/cmd/main.go
```
---
