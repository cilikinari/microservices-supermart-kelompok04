import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { Product, Category, Brand } from '../../../core/models/types';
import { CatalogService } from '../../../core/services/catalog.service';
import { CartService } from '../../../core/services/cart.service';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-catalog',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  template: `
    <div class="container catalog-page">
      <!-- Hero Banner -->
      <div class="hero-banner">
        <div class="hero-text">
          <span class="hero-badge">🛒 Belanja Pintar SuperMart</span>
          <h1>Katalog Produk Nusantara</h1>
          <p>Kebutuhan pokok, bahan pangan organik, alat kerja, dan perangkat rumah tangga terbaik.</p>
        </div>
        <div class="hero-stats">
          <div class="stat-pill">
            <span class="stat-num">120+</span>
            <span class="stat-lbl">Tabel Database</span>
          </div>
          <div class="stat-pill">
            <span class="stat-num">9</span>
            <span class="stat-lbl">Domain Cluster</span>
          </div>
          <div class="stat-pill">
            <span class="stat-num">100%</span>
            <span class="stat-lbl">Monolith Seam</span>
          </div>
        </div>
      </div>

      <!-- Informative Notice for Staff/Admin in Catalog -->
      @if (auth.isLoggedIn() && auth.userRole() !== 'CUSTOMER') {
        <div class="card mb-6" style="background: rgba(14, 165, 233, 0.08); border-left: 4px solid var(--color-primary);">
          <div class="flex items-center gap-3">
            <span style="font-size: 1.5rem;">ℹ️</span>
            <div>
              <strong style="color: var(--color-primary-dark);">Mode Tinjau Katalog (Peran: {{ auth.userRole() }})</strong>
              <p class="text-sm text-secondary mb-0">
                Fitur keranjang belanja dan pemesanan dinonaktifkan untuk staf/admin. Hanya akun Pelanggan yang dapat berbelanja.
              </p>
            </div>
          </div>
        </div>
      }

      <!-- Filter and Search Toolbar -->
      <div class="toolbar card">
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input
            type="text"
            class="search-input"
            [(ngModel)]="searchQuery"
            (input)="onSearchChange()"
            placeholder="Cari beras, kopi kintamani, mouse, keyboard..."
            id="input-search-catalog"
          />
        </div>

        <div class="filter-controls">
          <select class="select-control" [(ngModel)]="selectedCategory" (change)="loadProducts()" id="select-category">
            <option value="">Semua Kategori</option>
            @for (cat of categories(); track cat.id) {
              <option [value]="cat.id">{{ cat.category_name }}</option>
            }
          </select>

          <select class="select-control" [(ngModel)]="sortBy" (change)="loadProducts()" id="select-sort">
            <option value="">Urutkan Default</option>
            <option value="price_asc">Harga Terendah</option>
            <option value="price_desc">Harga Tertinggi</option>
            <option value="title_asc">Nama A-Z</option>
          </select>
        </div>
      </div>

      <!-- Categories Pills -->
      <div class="category-pills">
        <button
          class="cat-chip"
          [class.active]="selectedCategory === ''"
          (click)="setCategory('')"
        >
          Semua
        </button>
        @for (cat of categories(); track cat.id) {
          <button
            class="cat-chip"
            [class.active]="selectedCategory === cat.id"
            (click)="setCategory(cat.id)"
          >
            {{ cat.category_name }}
          </button>
        }
      </div>

      <!-- Product Grid -->
      @if (loading()) {
        <div class="loading-state">
          <span class="spinner">⏳</span>
          <p>Memuat katalog produk...</p>
        </div>
      } @else if (products().length === 0) {
        <div class="empty-state card">
          <span class="empty-icon">📦</span>
          <h3>Tidak ada produk yang cocok</h3>
          <p>Coba gunakan kata kunci pencarian atau kategori lain.</p>
        </div>
      } @else {
        <div class="product-grid">
          @for (product of products(); track product.id) {
            <div class="product-card card" [id]="'product-' + product.id">
              <div class="card-img-wrap cursor-pointer" [routerLink]="['/catalog', product.id]">
                <img
                  [src]="product.image_url || 'https://images.unsplash.com/photo-1586201375761-83865001e31c?w=500'"
                  [alt]="product.title"
                  class="product-img"
                />
                @if (product.brand_name) {
                  <span class="brand-badge">{{ product.brand_name }}</span>
                }
              </div>

              <div class="card-body">
                <span class="category-text">{{ product.category_name || 'Katalog Retail' }}</span>
                <h3 class="product-title cursor-pointer" [routerLink]="['/catalog', product.id]" [title]="product.title">{{ product.title }}</h3>
                <p class="product-desc">{{ product.description }}</p>

                <div class="product-meta">
                  <div class="price-wrap">
                    <span class="currency">Rp</span>
                    <span class="price">{{ product.base_price | number:'1.0-0' }}</span>
                  </div>
                  <div class="rating-wrap" title="Rating pembeli">
                    <span class="star">⭐</span>
                    <span class="rating-val">{{ product.rating | number:'1.1-1' }}</span>
                    <span class="review-count">({{ product.review_count }})</span>
                  </div>
                </div>

                <div class="card-footer">
                  <span class="weight-info">⚖️ {{ product.weight_gram }}g</span>
                  @if (!auth.isLoggedIn() || auth.userRole() === 'CUSTOMER') {
                    <button
                      (click)="addToCart(product)"
                      class="btn btn-primary btn-sm add-cart-btn"
                      [id]="'btn-add-cart-' + product.id"
                      [disabled]="addingToCart() === product.id"
                    >
                      <span>{{ addingToCart() === product.id ? 'Menambahkan...' : '+ Keranjang' }}</span>
                    </button>
                  }
                </div>

              </div>
            </div>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .catalog-page {
      padding-bottom: 3rem;
    }
    .hero-banner {
      background: linear-gradient(135deg, #0284c7 0%, #0369a1 50%, #0f172a 100%);
      color: white;
      border-radius: var(--radius-xl);
      padding: 2.5rem;
      margin-bottom: 2rem;
      display: flex;
      justify-content: space-between;
      align-items: center;
      box-shadow: var(--shadow-lg);
    }
    .hero-badge {
      display: inline-block;
      font-size: 0.75rem;
      font-weight: 700;
      background: rgba(255, 255, 255, 0.2);
      padding: 0.25rem 0.75rem;
      border-radius: var(--radius-full);
      margin-bottom: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .hero-text h1 {
      font-size: 2rem;
      font-weight: 800;
      letter-spacing: -0.02em;
    }
    .hero-text p {
      color: #e0f2fe;
      margin-top: 0.5rem;
      max-width: 600px;
    }
    .hero-stats {
      display: flex;
      gap: 1rem;
    }
    .stat-pill {
      background: rgba(255, 255, 255, 0.1);
      backdrop-filter: blur(8px);
      padding: 1rem 1.25rem;
      border-radius: var(--radius-lg);
      border: 1px solid rgba(255, 255, 255, 0.2);
      display: flex;
      flex-direction: column;
      align-items: center;
    }
    .stat-num {
      font-size: 1.5rem;
      font-weight: 800;
      color: #38bdf8;
    }
    .stat-lbl {
      font-size: 0.6875rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #e0f2fe;
    }
    .toolbar {
      padding: 1.25rem;
      margin-bottom: 1.5rem;
      display: flex;
      gap: 1rem;
      align-items: center;
      justify-content: space-between;
      flex-wrap: wrap;
    }
    .search-box {
      position: relative;
      flex: 1;
      min-width: 280px;
    }
    .search-icon {
      position: absolute;
      left: 0.875rem;
      top: 50%;
      transform: translateY(-50%);
      font-size: 1rem;
    }
    .search-input {
      width: 100%;
      padding: 0.625rem 1rem 0.625rem 2.5rem;
      font-size: 0.875rem;
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      background: var(--bg-surface-secondary);
      transition: all var(--transition-fast);
    }
    .search-input:focus {
      outline: none;
      border-color: var(--color-primary);
      background: white;
      box-shadow: 0 0 0 3px rgba(2, 132, 199, 0.15);
    }
    .filter-controls {
      display: flex;
      gap: 0.75rem;
    }
    .select-control {
      padding: 0.625rem 1rem;
      font-size: 0.875rem;
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      background: white;
      color: var(--text-primary);
      cursor: pointer;
    }
    .category-pills {
      display: flex;
      gap: 0.5rem;
      overflow-x: auto;
      padding-bottom: 1rem;
      margin-bottom: 1.5rem;
    }
    .cat-chip {
      padding: 0.4rem 1rem;
      font-size: 0.8125rem;
      font-weight: 600;
      border-radius: var(--radius-full);
      border: 1px solid var(--border-color);
      background: white;
      color: var(--text-secondary);
      cursor: pointer;
      white-space: nowrap;
      transition: all var(--transition-fast);
    }
    .cat-chip:hover {
      background: var(--bg-surface-secondary);
      color: var(--text-primary);
    }
    .cat-chip.active {
      background: var(--color-primary);
      color: white;
      border-color: var(--color-primary);
    }
    .product-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 1.5rem;
    }
    .product-card {
      display: flex;
      flex-direction: column;
      border-radius: var(--radius-lg);
      transition: transform var(--transition-fast), box-shadow var(--transition-fast);
    }
    .product-card:hover {
      transform: translateY(-4px);
      box-shadow: var(--shadow-lg);
    }
    .cursor-pointer {
      cursor: pointer;
    }
    .card-img-wrap {
      position: relative;
      width: 100%;
      height: 200px;
      background: var(--bg-surface-secondary);
      overflow: hidden;
    }
    .product-img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      transition: transform var(--transition-normal);
    }
    .product-card:hover .product-img {
      transform: scale(1.05);
    }
    .brand-badge {
      position: absolute;
      top: 0.75rem;
      left: 0.75rem;
      background: rgba(15, 23, 42, 0.75);
      backdrop-filter: blur(4px);
      color: white;
      font-size: 0.6875rem;
      font-weight: 700;
      padding: 0.2rem 0.5rem;
      border-radius: var(--radius-sm);
    }
    .card-body {
      padding: 1.25rem;
      display: flex;
      flex-direction: column;
      flex: 1;
    }
    .category-text {
      font-size: 0.6875rem;
      font-weight: 700;
      color: var(--color-primary);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .product-title {
      font-size: 1rem;
      font-weight: 700;
      color: var(--text-primary);
      margin: 0.25rem 0 0.5rem;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      min-height: 2.8rem;
    }
    .product-desc {
      font-size: 0.8125rem;
      color: var(--text-secondary);
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      margin-bottom: 1rem;
    }
    .product-meta {
      margin-top: auto;
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 1rem;
    }
    .price-wrap {
      display: flex;
      align-items: baseline;
      gap: 0.125rem;
    }
    .currency {
      font-size: 0.8125rem;
      font-weight: 700;
      color: var(--color-primary-dark);
    }
    .price {
      font-size: 1.25rem;
      font-weight: 800;
      color: var(--color-primary-dark);
    }
    .rating-wrap {
      display: flex;
      align-items: center;
      gap: 0.25rem;
      font-size: 0.75rem;
    }
    .rating-val {
      font-weight: 700;
      color: var(--text-primary);
    }
    .review-count {
      color: var(--text-muted);
    }
    .card-footer {
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-top: 1px solid var(--border-color);
      padding-top: 0.875rem;
    }
    .weight-info {
      font-size: 0.75rem;
      color: var(--text-muted);
      font-weight: 600;
    }
    .add-cart-btn {
      border-radius: var(--radius-md);
    }
    .loading-state, .empty-state {
      text-align: center;
      padding: 4rem 2rem;
    }
    .spinner {
      font-size: 2.5rem;
      display: inline-block;
      margin-bottom: 1rem;
      animation: spin 1s infinite linear;
    }
    .empty-icon {
      font-size: 3rem;
      display: inline-block;
      margin-bottom: 1rem;
    }
  `]
})
export class CatalogComponent implements OnInit {
  private catalogService = inject(CatalogService);
  private cartService = inject(CartService);
  auth = inject(AuthService);

  products = signal<Product[]>([]);
  categories = signal<Category[]>([]);
  loading = signal(true);
  addingToCart = signal<string | null>(null);

  searchQuery = '';
  selectedCategory = '';
  sortBy = '';

  ngOnInit() {
    this.loadCategories();
    this.loadProducts();
  }

  loadCategories() {
    this.catalogService.getCategories().subscribe(res => {
      if (res.success && res.data) {
        this.categories.set(res.data);
      }
    });
  }

  loadProducts() {
    this.loading.set(true);
    this.catalogService.getProducts({
      search: this.searchQuery || undefined,
      category_id: this.selectedCategory || undefined,
      sort_by: this.sortBy || undefined,
      limit: 24
    }).subscribe({
      next: res => {
        if (res.success && res.data) {
          this.products.set(res.data);
        }
        this.loading.set(false);
      },
      error: () => this.loading.set(false)
    });
  }

  setCategory(catId: string) {
    this.selectedCategory = catId;
    this.loadProducts();
  }

  onSearchChange() {
    this.loadProducts();
  }

  addToCart(product: Product) {
    if (this.auth.isLoggedIn() && this.auth.userRole() !== 'CUSTOMER') {
      return;
    }
    this.addingToCart.set(product.id);
    this.cartService.addToCart(product.id, 1).subscribe({
      next: () => {
        this.addingToCart.set(null);
      },
      error: () => {
        this.addingToCart.set(null);
      }
    });
  }
}

