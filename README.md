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

**Menyalakan semua basis data (berjalan di background):**
```bash
docker compose -f deployments/docker/docker-compose.yml up -d
```

**Mematikan semua basis data:**
```bash
docker compose -f deployments/docker/docker-compose.yml down
```