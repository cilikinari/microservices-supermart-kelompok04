-- =============================================================================
-- Order Service Database Initialization (order_service_db)
-- Designed for Database-per-Service Architecture & Polyglot Persistence
-- No physical cross-database foreign keys (stores logical ID references)
-- =============================================================================

CREATE DATABASE IF NOT EXISTS `order_service_db`;
USE `order_service_db`;

SET FOREIGN_KEY_CHECKS = 0;

-- 1. Status Pesanan (Order Statuses)
CREATE TABLE IF NOT EXISTS `order_statuses` (
    `id` VARCHAR(36) PRIMARY KEY,
    `status_code` VARCHAR(32) NOT NULL UNIQUE,
    `status_name` VARCHAR(100) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Keranjang Belanja (Carts)
CREATE TABLE IF NOT EXISTS `carts` (
    `id` VARCHAR(36) PRIMARY KEY,
    `user_id` VARCHAR(36) NOT NULL UNIQUE,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `cart_items` (
    `id` VARCHAR(36) PRIMARY KEY,
    `cart_id` VARCHAR(36) NOT NULL,
    `product_id` VARCHAR(36) NOT NULL,
    `quantity` INT NOT NULL DEFAULT 1,
    INDEX `idx_cart_id` (`cart_id`),
    INDEX `idx_product_id` (`product_id`),
    CONSTRAINT `fk_cart_items_cart` FOREIGN KEY (`cart_id`) REFERENCES `carts`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Transaksi Pesanan (Orders & Snapshots)
CREATE TABLE IF NOT EXISTS `orders` (
    `id` VARCHAR(36) PRIMARY KEY,
    `order_number` VARCHAR(64) NOT NULL UNIQUE,
    `customer_id` VARCHAR(36) NOT NULL,
    `shipping_address_id` VARCHAR(36) NOT NULL,
    `order_status_id` VARCHAR(36) NOT NULL,
    `warehouse_id` VARCHAR(36) NOT NULL,
    `total_gross_amount` DECIMAL(15,2) NOT NULL,
    `discount_amount` DECIMAL(15,2) DEFAULT 0.00,
    `tax_amount` DECIMAL(15,2) DEFAULT 0.00,
    `shipping_fee` DECIMAL(15,2) NOT NULL,
    `total_net_amount` DECIMAL(15,2) NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_customer_id` (`customer_id`),
    INDEX `idx_order_status_id` (`order_status_id`),
    CONSTRAINT `fk_orders_status` FOREIGN KEY (`order_status_id`) REFERENCES `order_statuses`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `order_items` (
    `id` VARCHAR(36) PRIMARY KEY,
    `order_id` VARCHAR(36) NOT NULL,
    `product_id` VARCHAR(36) NOT NULL,
    `product_title` VARCHAR(255) NULL,
    `product_sku` VARCHAR(100) NULL,
    `quantity` INT NOT NULL,
    `unit_price` DECIMAL(15,2) NOT NULL,
    `subtotal` DECIMAL(15,2) NOT NULL,
    INDEX `idx_order_id` (`order_id`),
    INDEX `idx_product_id` (`product_id`),
    CONSTRAINT `fk_order_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `order_shipping_details` (
    `id` VARCHAR(36) PRIMARY KEY,
    `order_id` VARCHAR(36) NOT NULL UNIQUE,
    `courier_name` VARCHAR(100) NOT NULL,
    `tracking_number` VARCHAR(100) NULL,
    `shipping_cost` DECIMAL(15,2) NOT NULL,
    `recipient_name` VARCHAR(150) NULL,
    `shipping_address` TEXT NULL,
    CONSTRAINT `fk_shipping_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `order_status_histories` (
    `id` VARCHAR(36) PRIMARY KEY,
    `order_id` VARCHAR(36) NOT NULL,
    `order_status_id` VARCHAR(36) NOT NULL,
    `notes` TEXT NULL,
    `changed_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_order_id` (`order_id`),
    CONSTRAINT `fk_history_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_history_status` FOREIGN KEY (`order_status_id`) REFERENCES `order_statuses`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `order_notes` (
    `id` VARCHAR(36) PRIMARY KEY,
    `order_id` VARCHAR(36) NOT NULL,
    `author_user_id` VARCHAR(36) NOT NULL,
    `note_content` TEXT NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_notes_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Finansial & Logistik Pendukung Transaksi
CREATE TABLE IF NOT EXISTS `payment_invoices` (
    `id` VARCHAR(36) PRIMARY KEY,
    `invoice_number` VARCHAR(64) NOT NULL UNIQUE,
    `order_id` VARCHAR(36) NOT NULL,
    `customer_id` VARCHAR(36) NOT NULL,
    `amount` DECIMAL(15,2) NOT NULL,
    `payment_status` VARCHAR(50) DEFAULT 'UNPAID',
    `due_date` DATETIME NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `shipping_orders` (
    `id` VARCHAR(36) PRIMARY KEY,
    `order_id` VARCHAR(36) NOT NULL,
    `courier_service_id` VARCHAR(36) NOT NULL,
    `tracking_number` VARCHAR(100) NOT NULL,
    `weight_kg` DECIMAL(6,2) NOT NULL,
    `current_status` VARCHAR(50) DEFAULT 'READY_FOR_PICKUP',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `shipping_tracking_logs` (
    `id` VARCHAR(36) PRIMARY KEY,
    `shipping_order_id` VARCHAR(36) NOT NULL,
    `status_description` VARCHAR(255) NOT NULL,
    `location` VARCHAR(150) NOT NULL,
    `logged_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_shipping_order_id` (`shipping_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Audit Log Pengujian RabbitMQ
CREATE TABLE IF NOT EXISTS `rabbitmq_processed_events` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `event_id` VARCHAR(64) NOT NULL UNIQUE,
    `event_type` VARCHAR(64) NOT NULL,
    `order_id` VARCHAR(64) NOT NULL,
    `product_id` VARCHAR(64) NOT NULL,
    `quantity` INT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Seed Data Status Order
INSERT IGNORE INTO `order_statuses` (`id`, `status_code`, `status_name`) VALUES
('os000001-0000-0000-0000-000000000001', 'AWAITING_PAYMENT', 'Menunggu Pembayaran'),
('os000001-0000-0000-0000-000000000002', 'PROCESSING', 'Pembayaran Lunas, Sedang Dikemas'),
('os000001-0000-0000-0000-000000000003', 'SHIPPED', 'Pesanan Telah Diserahkan ke Kurir'),
('os000001-0000-0000-0000-000000000004', 'COMPLETED', 'Pesanan Selesai Diterima'),
('os000001-0000-0000-0000-000000000005', 'CANCELLED', 'Pesanan Dibatalkan');

SET FOREIGN_KEY_CHECKS = 1;

