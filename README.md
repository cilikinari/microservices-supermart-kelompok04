# Dokumentasi Project Microservices Kelompok 04

## Tata Cara Instalasi dan Penggunaan Project
1. Cara cloning repositori
   Jalankan perintah berikut pada terminal anda:

   ```bash
   git clone <URL_REPOSITORY_GITHUB_TIM_ANDA>
   cd <NAMA_FOLDER_REPOSITORI>

2. Daftar Port, DBMS, & Kredensial Basis Data

| Domain | Teknologi DBMS | Port Host | Nama Database | Username | Password |
|---|---|---|---|---|---|
| **Identity** | MySQL 8.0 | `3306` | `identity_service_db` | `identity_admin` | `identity_secret_pass` |
| **Catalog** | MySQL 8.0 | `3307` | `catalog_service_db` | `catalog_admin` | `catalog_secret_pass` |
| **Inventory** | MySQL 8.0 | `3308` | `inventory_service_db` | `inventory_admin` | `inventory_secret_pass` |
| **Order** | MySQL 8.0 | `3309` | `order_service_db` | `order_admin` | `order_secret_pass` |

3. Perintah Menyalakan & Mematikan Lingkungan Kontainer
   - Menyalakan Kontainer

      ```bash
      docker-compose up -d 

      Jika menggunakan Docker Compose V2

      docker compose up -d
   
   - Mematikan Kontainer
      ```bash
      docker-compose down