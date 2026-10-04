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
