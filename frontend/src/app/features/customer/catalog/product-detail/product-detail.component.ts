import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterModule, Router } from '@angular/router';
import { CatalogService } from '../../../../core/services/catalog.service';
import { CartService } from '../../../../core/services/cart.service';
import { AuthService } from '../../../../core/services/auth.service';
import { NotificationService } from '../../../../core/services/notification.service';
import { Product, ProductImage, ProductVariant, ProductReview } from '../../../../core/models/types';

@Component({
  selector: 'app-product-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  template: `
    <div class="container product-detail-page">
      <!-- Breadcrumb & Back Button -->
      <div class="nav-back">
        <a routerLink="/catalog" class="back-link">
          ← Kembali ke Katalog
        </a>
      </div>

      @if (loading()) {
        <div class="card loading-state">
          <span class="spinner">⏳</span>
          <p>Memuat detail produk...</p>
        </div>
      } @else if (!product()) {
        <div class="card empty-state">
          <span class="empty-icon">❌</span>
          <h3>Produk Tidak Ditemukan</h3>
          <p>Produk yang Anda cari tidak tersedia atau sudah dihapus.</p>
          <a routerLink="/catalog" class="btn btn-primary mt-4">Lihat Katalog Lainnya</a>
        </div>
      } @else {
        <!-- Product Main Info Card -->
        <div class="card product-card-main">
          <div class="product-layout">
            <!-- Left: Gallery / Image -->
            <div class="gallery-section">
              <div class="main-image-wrap">
                <img
                  [src]="selectedImage() || product()?.image_url || 'https://images.unsplash.com/photo-1586201375761-83865001e31c?w=600'"
                  [alt]="product()?.title"
                  class="main-image"
                />
              </div>

              @if (images().length > 1) {
                <div class="thumbnails-wrap">
                  @for (img of images(); track img.id) {
                    <img
                      [src]="img.image_url"
                      [alt]="product()?.title"
                      class="thumb-img"
                      [class.active]="selectedImage() === img.image_url"
                      (click)="selectedImage.set(img.image_url)"
                    />
                  }
                </div>
              }
            </div>

            <!-- Right: Details & Order Controls -->
            <div class="details-section">
              <div class="badges-row">
                <span class="badge badge-info">{{ product()?.category_name || 'Katalog Retail' }}</span>
                @if (product()?.brand_name) {
                  <span class="badge badge-purple">{{ product()?.brand_name }}</span>
                }
              </div>

              <h1 class="product-title">{{ product()?.title }}</h1>
              <div class="sku-rating-row">
                <span class="sku-text font-mono">SKU: {{ product()?.sku }}</span>
                <span class="dot-divider">•</span>
                <div class="rating-box">
                  <span class="star">⭐</span>
                  <span class="rating-num">{{ product()?.rating | number:'1.1-1' }}</span>
                  <span class="review-count">({{ product()?.review_count }} ulasan)</span>
                </div>
              </div>

              <!-- Price & Weight -->
              <div class="price-box">
                <div class="price-val">
                  <span class="currency">Rp</span>
                  <span class="amount">{{ product()?.base_price | number:'1.0-0' }}</span>
                </div>
                <div class="weight-tag">
                  ⚖️ Berat: <strong>{{ product()?.weight_gram }} gram</strong>
                </div>
              </div>

              <!-- Description -->
              <div class="desc-box">
                <h3>Deskripsi Produk</h3>
                <p>{{ product()?.description || 'Tidak ada deskripsi rinci untuk produk ini.' }}</p>
              </div>

              <!-- Purchase Actions -->
              <div class="purchase-box">
                @if (auth.isLoggedIn() && auth.userRole() !== 'CUSTOMER') {
                  <div class="alert-notice">
                    ℹ️ <strong>Mode Tinjau (Peran: {{ auth.userRole() }})</strong>: Pembelian hanya dapat dilakukan oleh akun Pelanggan (Customer).
                  </div>
                } @else {
                  <div class="quantity-control-row">
                    <label class="qty-label">Jumlah Pembelian:</label>
                    <div class="qty-stepper">
                      <button (click)="decreaseQty()" class="qty-btn" [disabled]="quantity() <= 1">-</button>
                      <input type="number" [value]="quantity()" (change)="onQtyChange($event)" min="1" class="qty-input" />
                      <button (click)="increaseQty()" class="qty-btn">+</button>
                    </div>
                  </div>

                  <div class="action-buttons">
                    <button
                      (click)="addToCart()"
                      class="btn btn-primary btn-lg"
                      [disabled]="addingToCart()"
                    >
                      <span>{{ addingToCart() ? 'Menambahkan...' : '🛒 Masukkan ke Keranjang' }}</span>
                    </button>
                  </div>
                }
              </div>
            </div>
          </div>
        </div>

        <!-- Customer Reviews Section -->
        <div class="card reviews-section mt-6">
          <div class="section-header">
            <h2>Ulasan Pelanggan</h2>
            <span class="review-total">{{ reviews().length }} Ulasan</span>
          </div>

          <!-- Add Review Form (Only for logged-in CUSTOMER) -->
          @if (auth.isLoggedIn() && auth.userRole() === 'CUSTOMER') {
            <div class="add-review-card">
              <h4>Tulis Ulasan Anda</h4>
              <form (ngSubmit)="submitReview()">
                <div class="rating-select-row mb-3">
                  <label class="form-label">Beri Bintang:</label>
                  <div class="star-rating-picker">
                    @for (star of [1, 2, 3, 4, 5]; track star) {
                      <button
                        type="button"
                        class="star-pick-btn"
                        [class.active]="newRating >= star"
                        (click)="newRating = star"
                      >
                        ⭐
                      </button>
                    }
                    <span class="rating-label-text">({{ newRating }} dari 5)</span>
                  </div>
                </div>

                <div class="form-group mb-3">
                  <label class="form-label">Komentar / Pengalaman:</label>
                  <textarea
                    [(ngModel)]="newReviewText"
                    name="review_text"
                    rows="3"
                    class="form-control"
                    placeholder="Ceritakan kualitas bahan pangan atau produk ini..."
                  ></textarea>
                </div>

                <button type="submit" class="btn btn-secondary btn-sm" [disabled]="submittingReview()">
                  {{ submittingReview() ? 'Mengirim...' : 'Kirim Ulasan' }}
                </button>
              </form>
            </div>
          }

          <!-- Reviews List -->
          @if (loadingReviews()) {
            <p class="text-secondary py-4">Memuat ulasan produk...</p>
          } @else if (reviews().length === 0) {
            <div class="no-reviews py-6 text-center text-secondary">
              <p>Belum ada ulasan untuk produk ini. Jadilah yang pertama memberikan ulasan!</p>
            </div>
          } @else {
            <div class="reviews-list">
              @for (rev of reviews(); track rev.id) {
                <div class="review-item">
                  <div class="review-header">
                    <span class="reviewer-name font-bold">{{ rev.user_name || 'Pelanggan SuperMart' }}</span>
                    <span class="review-date text-xs text-muted">{{ rev.created_at | date:'dd MMM yyyy, HH:mm' }}</span>
                  </div>
                  <div class="review-stars mb-2">
                    @for (s of getStarArray(rev.rating); track $index) {
                      <span>⭐</span>
                    }
                  </div>
                  @if (rev.review_text) {
                    <p class="review-comment text-secondary">{{ rev.review_text }}</p>
                  }
                </div>
              }
            </div>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .product-detail-page {
      padding-bottom: 4rem;
    }
    .nav-back {
      margin-bottom: 1.5rem;
    }
    .back-link {
      font-weight: 600;
      color: var(--color-primary);
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
    }
    .product-layout {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 2.5rem;
    }
    @media (max-width: 768px) {
      .product-layout {
        grid-template-columns: 1fr;
      }
    }
    .gallery-section {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }
    .main-image-wrap {
      width: 100%;
      height: 380px;
      border-radius: var(--radius-lg);
      overflow: hidden;
      background: var(--bg-surface-secondary);
      display: flex;
      align-items: center;
      justify-content: center;
      border: 1px solid var(--border-color);
    }
    .main-image {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
    .thumbnails-wrap {
      display: flex;
      gap: 0.75rem;
      overflow-x: auto;
    }
    .thumb-img {
      width: 70px;
      height: 70px;
      border-radius: var(--radius-md);
      object-fit: cover;
      cursor: pointer;
      border: 2px solid transparent;
      transition: all var(--transition-fast);
    }
    .thumb-img:hover, .thumb-img.active {
      border-color: var(--color-primary);
    }
    .details-section {
      display: flex;
      flex-direction: column;
    }
    .badges-row {
      display: flex;
      gap: 0.5rem;
      margin-bottom: 0.75rem;
    }
    .product-title {
      font-size: 1.75rem;
      font-weight: 800;
      color: var(--text-primary);
      line-height: 1.3;
      margin-bottom: 0.75rem;
    }
    .sku-rating-row {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-bottom: 1.5rem;
      font-size: 0.875rem;
    }
    .dot-divider {
      color: var(--text-muted);
    }
    .rating-box {
      display: flex;
      align-items: center;
      gap: 0.25rem;
    }
    .rating-num {
      font-weight: 700;
      color: var(--text-primary);
    }
    .review-count {
      color: var(--text-muted);
    }
    .price-box {
      background: var(--bg-surface-secondary);
      padding: 1.25rem 1.5rem;
      border-radius: var(--radius-lg);
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
    }
    .price-val {
      display: flex;
      align-items: baseline;
      gap: 0.25rem;
    }
    .currency {
      font-size: 1.125rem;
      font-weight: 700;
      color: var(--color-primary-dark);
    }
    .amount {
      font-size: 2rem;
      font-weight: 800;
      color: var(--color-primary-dark);
    }
    .weight-tag {
      font-size: 0.875rem;
      color: var(--text-secondary);
    }
    .desc-box {
      margin-bottom: 2rem;
    }
    .desc-box h3 {
      font-size: 1rem;
      font-weight: 700;
      margin-bottom: 0.5rem;
    }
    .desc-box p {
      color: var(--text-secondary);
      line-height: 1.6;
      font-size: 0.9375rem;
    }
    .purchase-box {
      margin-top: auto;
      border-top: 1px solid var(--border-color);
      padding-top: 1.5rem;
    }
    .quantity-control-row {
      display: flex;
      align-items: center;
      gap: 1rem;
      margin-bottom: 1.25rem;
    }
    .qty-label {
      font-weight: 600;
      font-size: 0.875rem;
    }
    .qty-stepper {
      display: flex;
      align-items: center;
      border: 1px solid var(--border-color);
      border-radius: var(--radius-md);
      overflow: hidden;
      background: white;
    }
    .qty-btn {
      width: 38px;
      height: 38px;
      background: var(--bg-surface-secondary);
      border: none;
      font-size: 1.125rem;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: background var(--transition-fast);
    }
    .qty-btn:hover:not([disabled]) {
      background: var(--border-color);
    }
    .qty-input {
      width: 50px;
      height: 38px;
      border: none;
      text-align: center;
      font-weight: 700;
      font-size: 0.9375rem;
    }
    .qty-input:focus {
      outline: none;
    }
    .alert-notice {
      background: var(--color-primary-light);
      padding: 1rem;
      border-radius: var(--radius-md);
      font-size: 0.875rem;
      color: var(--color-primary-dark);
    }
    .reviews-section {
      padding: 2rem;
    }
    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
      border-bottom: 1px solid var(--border-color);
      padding-bottom: 0.75rem;
    }
    .review-total {
      background: var(--bg-surface-secondary);
      padding: 0.25rem 0.75rem;
      border-radius: var(--radius-full);
      font-size: 0.8125rem;
      font-weight: 700;
      color: var(--text-secondary);
    }
    .add-review-card {
      background: var(--bg-surface-secondary);
      padding: 1.25rem;
      border-radius: var(--radius-lg);
      margin-bottom: 2rem;
    }
    .add-review-card h4 {
      margin-bottom: 1rem;
      font-size: 1rem;
    }
    .star-rating-picker {
      display: flex;
      align-items: center;
      gap: 0.25rem;
    }
    .star-pick-btn {
      background: transparent;
      border: none;
      font-size: 1.5rem;
      cursor: pointer;
      opacity: 0.35;
      transition: opacity var(--transition-fast);
    }
    .star-pick-btn.active {
      opacity: 1;
    }
    .rating-label-text {
      margin-left: 0.5rem;
      font-size: 0.875rem;
      font-weight: 600;
      color: var(--text-secondary);
    }
    .review-item {
      padding: 1rem 0;
      border-bottom: 1px solid var(--border-color);
    }
    .review-item:last-child {
      border-bottom: none;
    }
    .review-header {
      display: flex;
      justify-content: space-between;
      margin-bottom: 0.25rem;
    }
    .spinner {
      font-size: 2.5rem;
      display: inline-block;
      margin-bottom: 1rem;
    }
    .loading-state, .empty-state {
      text-align: center;
      padding: 4rem 2rem;
    }
  `]
})
export class ProductDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private catalogService = inject(CatalogService);
  private cartService = inject(CartService);
  private notify = inject(NotificationService);
  auth = inject(AuthService);

  product = signal<Product | null>(null);
  images = signal<ProductImage[]>([]);
  variants = signal<ProductVariant[]>([]);
  reviews = signal<ProductReview[]>([]);
  selectedImage = signal<string | null>(null);

  loading = signal(true);
  loadingReviews = signal(false);
  submittingReview = signal(false);
  addingToCart = signal(false);
  quantity = signal(1);

  newRating = 5;
  newReviewText = '';

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.loadProduct(id);
      this.loadReviews(id);
    } else {
      this.loading.set(false);
    }
  }

  loadProduct(id: string) {
    this.loading.set(true);
    this.catalogService.getProduct(id).subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.product.set(res.data.product);
          this.images.set(res.data.images || []);
          this.variants.set(res.data.variants || []);
          if (res.data.images && res.data.images.length > 0) {
            this.selectedImage.set(res.data.images[0].image_url);
          }
        }
        this.loading.set(false);
      },
      error: () => {
        this.notify.error('Gagal memuat detail produk');
        this.loading.set(false);
      }
    });
  }

  loadReviews(productId: string) {
    this.loadingReviews.set(true);
    this.catalogService.getReviews(productId).subscribe({
      next: (res) => {
        if (res.success && res.data) {
          this.reviews.set(res.data);
        }
        this.loadingReviews.set(false);
      },
      error: () => {
        this.loadingReviews.set(false);
      }
    });
  }

  increaseQty() {
    this.quantity.update(q => q + 1);
  }

  decreaseQty() {
    this.quantity.update(q => (q > 1 ? q - 1 : 1));
  }

  onQtyChange(event: Event) {
    const val = Number((event.target as HTMLInputElement).value);
    if (!isNaN(val) && val >= 1) {
      this.quantity.set(val);
    } else {
      this.quantity.set(1);
    }
  }

  addToCart() {
    const p = this.product();
    if (!p) return;

    if (this.auth.isLoggedIn() && this.auth.userRole() !== 'CUSTOMER') {
      return;
    }

    this.addingToCart.set(true);
    this.cartService.addToCart(p.id, this.quantity()).subscribe({
      next: () => {
        this.notify.success(`${this.quantity()}x ${p.title} berhasil ditambahkan ke keranjang!`);
        this.addingToCart.set(false);
      },
      error: () => {
        this.notify.error('Gagal menambahkan produk ke keranjang');
        this.addingToCart.set(false);
      }
    });
  }

  submitReview() {
    const p = this.product();
    if (!p) return;

    if (this.newRating < 1 || this.newRating > 5) {
      this.notify.warning('Pilih rating 1 sampai 5');
      return;
    }

    this.submittingReview.set(true);
    this.catalogService.addReview(p.id, this.newRating, this.newReviewText).subscribe({
      next: () => {
        this.notify.success('Ulasan Anda berhasil dikirim!');
        this.newReviewText = '';
        this.newRating = 5;
        this.submittingReview.set(false);
        this.loadReviews(p.id);
      },
      error: (err) => {
        this.notify.error(err.error?.message || 'Gagal mengirim ulasan');
        this.submittingReview.set(false);
      }
    });
  }

  getStarArray(rating: number): number[] {
    const r = Math.max(1, Math.min(5, Math.round(rating || 5)));
    return Array(r).fill(0);
  }
}

