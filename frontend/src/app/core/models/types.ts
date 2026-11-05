export interface User {
  id: string;
  email: string;
  full_name: string;
  phone_number?: string;
  is_active: boolean;
  role?: string;
  created_at?: string;
}

export interface UserProfile {
  id: string;
  user_id: string;
  gender?: string;
  birth_date?: string;
  avatar_url?: string;
  bio?: string;
}

export interface UserAddress {
  id: string;
  user_id: string;
  address_label: string;
  recipient_name: string;
  phone_number: string;
  street_address: string;
  city: string;
  province: string;
  postal_code: string;
  is_primary: boolean;
}

export interface Product {
  id: string;
  sku: string;
  title: string;
  description?: string;
  base_price: number;
  brand_id?: string;
  weight_gram: number;
  is_published: boolean;
  brand_name?: string;
  category_name?: string;
  category_id?: string;
  image_url?: string;
  rating: number;
  review_count: number;
}

export interface Category {
  id: string;
  category_name: string;
  slug: string;
  icon_url?: string;
}

export interface Brand {
  id: string;
  brand_name: string;
  logo_url?: string;
  website_url?: string;
}

export interface ProductVariant {
  id: string;
  product_id: string;
  variant_sku: string;
  additional_price: number;
}

export interface ProductImage {
  id: string;
  product_id: string;
  image_url: string;
  is_primary: boolean;
  display_order: number;
}

export interface ProductDetailData {
  product: Product;
  variants: ProductVariant[];
  images: ProductImage[];
}

export interface ProductReview {
  id: string;
  product_id: string;
  user_id: string;
  rating: number;
  review_text?: string;
  created_at: string;
  user_name?: string;
}

export interface CartItem {
  id: string;
  cart_id: string;
  product_id: string;
  quantity: number;
  product_title: string;
  product_sku: string;
  product_price: number;
  image_url: string;
  subtotal: number;
}

export interface Cart {
  id: string;
  customer_id: string;
  items?: CartItem[];
}

export interface Order {
  id: string;
  order_number: string;
  customer_id: string;
  shipping_address_id: string;
  order_status_id: string;
  warehouse_id: string;
  total_gross_amount: number;
  discount_amount: number;
  tax_amount: number;
  shipping_fee: number;
  total_net_amount: number;
  created_at: string;
  updated_at: string;
  status_code?: string;
  status_name?: string;
  customer_name?: string;
  customer_email?: string;
  warehouse_name?: string;
  shipping_address?: string;
}

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  quantity: number;
  unit_price: number;
  subtotal: number;
  product_title?: string;
  product_sku?: string;
}

export interface OrderShippingDetail {
  id: string;
  order_id: string;
  courier_name: string;
  tracking_number?: string;
  shipping_cost: number;
}

export interface OrderStatusHistory {
  id: string;
  order_id: string;
  order_status_id: string;
  notes?: string;
  created_at: string;
  status_name?: string;
}

export interface OrderStatus {
  id: string;
  status_code: string;
  status_name: string;
}

export interface PaymentInvoice {
  id: string;
  invoice_number: string;
  order_id: string;
  customer_id: string;
  amount: number;
  payment_status: string; // UNPAID, PAID, EXPIRED, CANCELLED
  due_date: string;
  paid_at?: string;
  order_number?: string;
}

export interface PaymentMethod {
  id: string;
  method_code: string;
  method_name: string;
}

export interface StoreCredit {
  id: string;
  user_id: string;
  balance: number;
}

export interface CreditTransaction {
  id: string;
  store_credit_id: string;
  amount: number;
  transaction_type: string;
  description?: string;
  created_at: string;
}

export interface Stock {
  id: string;
  product_id: string;
  warehouse_id: string;
  quantity_on_hand: number;
  quantity_reserved: number;
  product_title: string;
  product_sku: string;
  product_price: number;
  warehouse_name: string;
  warehouse_city: string;
}

export interface Warehouse {
  id: string;
  warehouse_code: string;
  warehouse_name: string;
  address: string;
  city: string;
}

export interface StockMutation {
  id: string;
  product_id: string;
  source_warehouse_id: string;
  destination_warehouse_id: string;
  quantity: number;
  mutation_date: string;
  product_title?: string;
  source_warehouse_name?: string;
  destination_warehouse_name?: string;
}

export interface LowStockAlert {
  id: string;
  product_id: string;
  warehouse_id: string;
  current_stock: number;
  threshold: number;
  is_resolved: boolean;
  alert_time: string;
  product_title?: string;
  warehouse_name?: string;
}

export interface Voucher {
  id: string;
  voucher_code: string;
  voucher_value: number;
  quota_limit: number;
  quota_used: number;
  expires_at: string;
}

export interface LoyaltyPoint {
  id: string;
  user_id: string;
  current_points: number;
}

export interface ShippingOrder {
  id: string;
  order_id: string;
  courier_service_id: string;
  tracking_number: string;
  weight_kg: number;
  current_status: string;
  order_number?: string;
  service_name?: string;
  customer_name?: string;
  city?: string;
}

export interface Supplier {
  id: string;
  supplier_code: string;
  company_name: string;
  tax_identification_number: string;
  address: string;
}

export interface PurchaseOrder {
  id: string;
  po_number: string;
  supplier_id: string;
  warehouse_id: string;
  po_date: string;
  total_po_amount: number;
  status: string;
  supplier_name?: string;
  warehouse_name?: string;
}

export interface PurchaseOrderItem {
  id: string;
  purchase_order_id: string;
  product_id: string;
  ordered_qty: number;
  unit_cost: number;
  product_title?: string;
  product_sku?: string;
}

export interface CustomerTicket {
  id: string;
  ticket_code: string;
  customer_id: string;
  order_id?: string;
  category_id: string;
  subject: string;
  priority: string;
  status: string;
  created_at: string;
  customer_name?: string;
  category_name?: string;
  order_number?: string;
  agent_name?: string;
}

export interface TicketMessage {
  id: string;
  ticket_id: string;
  sender_user_id: string;
  message_body: string;
  sent_at: string;
  sender_name?: string;
  sender_role?: string;
}

export interface FAQArticle {
  id: string;
  category_id: string;
  question: string;
  answer: string;
  category_name?: string;
}

export interface GoodsReceiptNote {
  id: string;
  grn_number: string;
  purchase_order_id: string;
  received_by_user_id: string;
  receipt_date: string;
  po_number?: string;
  receiver_name?: string;
}

export interface Promotion {
  id: string;
  promo_name: string;
  description: string;
  start_date: string;
  end_date: string;
  is_active: boolean;
}

export interface UserKYC {
  id: string;
  user_id: string;
  id_card_number: string;
  full_name: string;
  verification_status: string; // PENDING, APPROVED, REJECTED
  submitted_at: string;
  user_email?: string;
}

export interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
  meta?: any;
}

export interface CourierService {
  id: string;
  courier_id: string;
  service_code: string;
  service_name: string;
  estimated_days?: string;
}

export interface TicketCategory {
  id: string;
  category_name: string;
  description?: string;
}


