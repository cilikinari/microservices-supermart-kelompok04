# 🛒 PT Nusantara SuperMart Indonesia - Microservices Platform
> **Tugas Besar Arsitektur Sistem Terdistribusi & Mikroservis**  
> **Kelompok 04**

---

### Rincian Basis Data Terisolasi

| Service | DBMS | Port Host | Nama Database | Username | Password |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Identity Service** | MySQL 8.0 | `3309` | `identity_service_db` | `identity_admin` | `identity_secret_pass` |
| **Catalog Service** | MySQL 8.0 | `3307` | `catalog_service_db` | `catalog_admin` | `catalog_secret_pass` |
| **Inventory Service**| MySQL 8.0 | `3308` | `inventory_service_db` | `inventory_admin` | `inventory_secret_pass` |
| **Order Service** | MySQL 8.0 | `3310` | `order_service_db` | `order_admin` | `order_secret_pass` |
| **RabbitMQ Broker** | 3.13 Mgmt | `5672` (AMQP), `15672` (UI) | Virtual Host: `/` | `supermart` | `supermart123` |

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

### 4. Verifikasi Kesehatan 7 Kontainer
Periksa status kontainer:
```bash
docker compose -f deployments/docker/docker-compose.yml ps
```
Semua 7 kontainer berikut **wajib berstatus `Up` / `healthy`**:
1. `supermart-identity-db`
2. `supermart-catalog-db`
3. `supermart-inventory-db`
4. `supermart-order-db`
5. `supermart-rabbitmq`
6. `supermart-backend`
7. `supermart-frontend`

### 5. Memantau Log Server
Untuk melihat proses inisialisasi gRPC, database, dan RabbitMQ consumer:
```bash
docker compose -f deployments/docker/docker-compose.yml logs -f backend
```
---

## 🧪 Skenario Pengujian Lengkap

### Skenario A: Uji Liveness & Healthcheck (REST API)
Buka di peramban atau jalankan lewat terminal:
```bash
curl http://localhost:3000/api/v1/health
```
**Ekspektasi Output:**
```json
{
  "success": true,
  "message": "Nusantara SuperMart Monolith API is operational",
  "data": {
    "database": "healthy",
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

Gunakan tool pengujian otomatis yang telah disediakan di `backend/cmd/publish-test`:

#### 1. Uji Pengiriman Massal (Batch Mode)
Menguji pengiriman 5 antrean transaksi secara asinkron:
```bash
cd backend
go run ./cmd/publish-test -mode=batch -count=5
```
* **Hasil di Log Backend:** Consumer menerima 5 event pesanan, memotong stok secara aman, dan mengirim sinyal `Ack`.

#### 2. Uji Idempotensi (Mitigasi Double-Reservation)
Menguji apa yang terjadi jika jaringan mengirimkan pesan pesanan yang sama 2 kali:
```bash
cd backend
go run ./cmd/publish-test -mode=duplicate -event-id=DEMO-DUP-001
```
* **Hasil di Log Backend:**  
  * Pesan 1: Diproses dan stok dikunci (*quantity_reserved + 2*).  
  * Pesan 2: Terdeteksi duplikat! Consumer mencetak log:  
    `[Idempotency] Event DEMO-DUP-001 ... sudah pernah diproses. Melewati mutasi stok ganda.`  

#### 3. Uji Poison Pill & Dead-Letter Queue (DLQ)
Menguji ketahanan sistem jika menerima pesan korup atau JSON cacat kontrak:
```bash
cd backend
go run ./cmd/publish-test -mode=poison
```
* **Hasil:** Consumer otomatis menolak pesan via `NACK (requeue=false)` dan RabbitMQ mengalihkannya ke **DLQ** (`inventory_stock_reservation_queue.dlq`) tanpa membuat worker backend crash! Cek antrean DLQ di browser [http://localhost:15672/#/queues](http://localhost:15672/#/queues).

---

### Skenario C: Verifikasi Komunikasi gRPC (Bagian Astrid)
Layanan gRPC otomatis aktif di background backend pada dua port:
* **Inventory gRPC**: Port `50051`
* **Catalog gRPC**: Port `50052`

Saat Frontend membuka etalase katalog atau saat k6 dijalankan, komunikasi gRPC langsung terjadi dan tercetak di log backend:
```bash
docker compose -f deployments/docker/docker-compose.yml logs -f backend
```
Log akan mencatat panggilan gRPC secara transparan (*real-time*). Untuk pengujian beban gRPC, cukup jalankan skrip k6 di Skenario E.

---

### Skenario D: Uji Antarmuka Web Frontend
Buka browser di:
* **URL Utama**: [http://localhost:4200](http://localhost:4200) atau [http://localhost:80](http://localhost:80)

1. **Jelajah Katalog**: Buka halaman produk. Frontend mengambil data dari backend via endpoint `/api/v1/catalog/products`.
2. **Login Akun Demo**:
   * Email: `wayan.putra@gmail.com`
   * Password: `password123`
3. **Simulasi Checkout**:
   * Masukkan produk ke keranjang belanja.
   * Lakukan checkout pesanan.
   * Amati terminal: Event `OrderCreated` otomatis terbit ke RabbitMQ dan stok gudang otomatis ter-reservasi di database Inventory.

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

> **Tips:** Jika k6 belum terpasang di host, kamu bisa menjalankannya langsung via Docker:  
> `docker run --rm --net=host -v $(pwd):/app -w /app grafana/k6 run scripts/grpc_test.js`

---

## 🛠️ Perintah Berguna (*Cheat Sheet*)

```bash
# 1. Menyalakan semua kontainer:
docker compose -f deployments/docker/docker-compose.yml up --build -d

# 2. Menghentikan semua kontainer:
docker compose -f deployments/docker/docker-compose.yml down

# 3. Menghentikan kontainer dan MERESET SEMUA DATA DATABASE (Fresh Start):
docker compose -f deployments/docker/docker-compose.yml down -v

# 4. Menjalankan Backend secara lokal (Host) untuk debugging:
# Matikan backend container terlebih dahulu agar port 3000 bebas:
docker compose -f deployments/docker/docker-compose.yml stop backend
# Jalankan backend lokal:
cd backend && go run main.go
```
---