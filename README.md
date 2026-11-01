# Implementasi Microservices SuperMart (Klp 4)

## 1. Cara Mengkloning Repositori
Jalankan perintah berikut pada terminal:
```bash
git clone [https://github.com/cilikinari/microservices-supermart.git](https://github.com/cilikinari/microservices-supermart.git)
cd microservices-supermart
```

## 2. Konfigurasi Basis Data per Domain
Berikut adalah daftar layanan beserta spesifikasi basis datanya:

| Domain Layanan | DBMS Pilihan | Port | Nama Database | Username | Password | Root Password |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Identity** | MySQL 8.0 | `3306` | `identity_service_db` | `identity_admin` | `identity_secret_pass` | `identity_pass` |
| **Catalog** | MySQL 8.0 | `3307` | `catalog_service_db` | `catalog_admin` | `catalog_secret_pass` | `root_catalog_pass` |
| **Inventory** | MySQL 8.0 | `3308` | `inventory_service_db` | `inventory_admin` | `inventory_secret_pass` | `root_inventory_pass` |

## 3. Menjalankan Lingkungan Kontainer
Pastikan Anda berada di direktori *root* proyek.

Backend membaca tiga koneksi database berikut dari `.env`:

```env
IDENTITY_DB_HOST=127.0.0.1
IDENTITY_DB_PORT=3306
IDENTITY_DB_NAME=identity_service_db
IDENTITY_DB_USER=identity_admin
IDENTITY_DB_PASSWORD=identity_secret_pass

CATALOG_DB_HOST=127.0.0.1
CATALOG_DB_PORT=3307
CATALOG_DB_NAME=catalog_service_db
CATALOG_DB_USER=catalog_admin
CATALOG_DB_PASSWORD=catalog_secret_pass

INVENTORY_DB_HOST=127.0.0.1
INVENTORY_DB_PORT=3308
INVENTORY_DB_NAME=inventory_service_db
INVENTORY_DB_USER=inventory_admin
INVENTORY_DB_PASSWORD=inventory_secret_pass
```

Pastikan database container sudah berjalan sebelum menjalankan backend. Domain
identity, catalog, dan inventory memakai koneksi masing-masing. Domain lain
belum memiliki database terpisah pada konfigurasi saat ini.

**Menyalakan semua basis data (berjalan di background):**
```bash
docker compose -f deployments/docker/docker-compose.yml up -d
```

**Mematikan semua basis data:**
```bash
docker compose -f deployments/docker/docker-compose.yml down
```


## Rekayasa Multi-Stage Build & Skrip Pengujian Kontainer 

Cabang `feat/docker-optimization` mengimplementasikan pola *Multi-Stage Build* dan *Container Hardening* pada Backend Go (`inventory-service`) dan Frontend Angular.

### Skrip Build Image

Jalankan perintah berikut dari direktori root untuk membangun seluruh varian image:

#### 1. Backend Service (Go)
```bash
# Baseline (Single-Stage Naive Build)
docker build -t supermart-backend:baseline -f backend/Dockerfile.baseline backend

# Multi-Stage Varian 1 (Alpine 3.19)
docker build -t supermart-backend:alpine -f backend/Dockerfile.alpine backend

# Multi-Stage Varian 2 (Google Distroless - Non-Root & No Shell)
docker build -t supermart-backend:distroless -f backend/Dockerfile.distroless backend

# Multi-Stage Varian 3 (Scratch - 0 Byte Base)
docker build -t supermart-backend:scratch -f backend/Dockerfile.scratch backend
```

#### 2. Frontend Web (Angular)
```bash
# Baseline (Single-Stage Dev Server)
docker build -t supermart-frontend:baseline -f frontend/Dockerfile.baseline frontend

# Multi-Stage Production (Nginx Alpine Serving)
docker build -t supermart-frontend:optimized frontend
```

---

### Skrip Menjalankan Kontainer Pengujian

```bash
# Jalankan Kontainer Backend (Pilih salah satu port varian)
docker run -d -p 3001:3000 -p 50051:50051 --name be-test-alpine supermart-backend:alpine
docker run -d -p 3002:3000 -p 50052:50051 --name be-test-distroless supermart-backend:distroless
docker run -d -p 3003:3000 -p 50053:50051 --name be-test-scratch supermart-backend:scratch

# Jalankan Kontainer Frontend (dengan mock upstream backend)
docker run -d -p 8080:80 --add-host backend:127.0.0.1 --name test-fe supermart-frontend:optimized
```

---

### Skrip Pengujian Operasionalitas (Health Check & Liveness Probe)

Perintah validasi untuk membuktikan bahwa seluruh kontainer melayani permintaan HTTP secara normal (`HTTP/1.1 200 OK`):

```bash
# 1. Uji Backend Varian A (Alpine)
curl.exe -i http://localhost:3001/api/v1/health

# 2. Uji Backend Varian B (Distroless)
curl.exe -i http://localhost:3002/api/v1/health

# 3. Uji Backend Varian C (Scratch)
curl.exe -i http://localhost:3003/api/v1/health

# 4. Uji Keamanan Hardening Distroless (Verifikasi Ketiadaan Shell / No Shell)
docker exec -it be-test-distroless /bin/sh
# Ekspektasi output: OCI runtime exec failed ... stat /bin/sh: no such file or directory

# 5. Uji Frontend Web Angular
curl.exe -i http://localhost:8080
# Atau buka peramban di: http://localhost:8080
```

---