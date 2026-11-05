import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Product, Category, Brand } from '../../../core/models/types';
import { CatalogService } from '../../../core/services/catalog.service';
import { NotificationService } from '../../../core/services/notification.service';

@Component({
  selector: 'app-admin-catalog-mgmt',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="container catalog-mgmt-page">
      <div class="page-header">
        <div>
          <span class="badge badge-purple">Super Admin Console</span>
          <h1 class="page-title">Kelola Katalog Produk & Master Data</h1>
          <p class="text-secondary">Tambah SKU produk retail, tentukan harga dasar, atur kategori dan merek.</p>
        </div>
        <div class="header-actions">
          <button (click)="openAddModal()" class="btn btn-primary">
            ➕ Tambah Produk Baru
          </button>
          <button (click)="loadProducts()" class="btn btn-outline" [disabled]="loading()">
            🔄 Refresh
          </button>
        </div>
      </div>

      <!-- Products Table Card -->
      <div class="card p-0">
        <div class="table-toolbar p-4 border-b flex justify-between items-center">
          <div class="search-box">
            <input 
              type="text" 
              [(ngModel)]="searchQuery" 
              placeholder="Cari nama produk, SKU, atau kategori..." 
              class="form-control"
            />
          </div>
          <span class="text-sm text-secondary">
            Menampilkan {{ filteredProducts().length }} dari {{ products().length }} produk
          </span>
        </div>

        @if (loading()) {
          <div class="p-8 text-center">
            <div class="spinner"></div>
            <p class="mt-2 text-secondary">Memuat katalog master...</p>
          </div>
        } @else if (filteredProducts().length === 0) {
          <div class="p-8 text-center text-secondary">
            <p>Tidak ada produk yang ditemukan.</p>
          </div>
        } @else {
          <div class="table-responsive">
            <table class="table">
              <thead>
                <tr>
                  <th>Gambar</th>
                  <th>Produk & SKU</th>
                  <th>Kategori</th>
                  <th>Merek (Brand)</th>
                  <th class="text-right">Harga Dasar</th>
                  <th class="text-right">Berat (gram)</th>
                  <th>Status Tayang</th>
                  <th class="text-right">Aksi</th>
                </tr>
              </thead>
              <tbody>
                @for (p of filteredProducts(); track p.id) {
                  <tr>
                    <td>
                      <img 
                        [src]="p.image_url || 'https://images.unsplash.com/photo-1542838132-92c53300491e?w=100&auto=format&fit=crop'" 
                        alt="{{ p.title }}" 
                        class="product-thumb"
                      />
                    </td>
                    <td>
                      <div class="product-info-cell">
                        <span class="font-bold text-primary">{{ p.title }}</span>
                        <span class="font-mono text-xs text-secondary">{{ p.sku }}</span>
                      </div>
                    </td>
                    <td>
                      <span class="badge badge-info">{{ p.category_name || 'Umum' }}</span>
                    </td>
                    <td>
                      <span class="font-semibold text-secondary">{{ p.brand_name || '-' }}</span>
                    </td>
                    <td class="text-right font-bold">
                      Rp {{ p.base_price | number:'1.0-0' }}
                    </td>
                    <td class="text-right text-secondary">{{ p.weight_gram }} g</td>
                    <td>
                      @if (p.is_published) {
                        <span class="badge badge-success">PUBLISHED</span>
                      } @else {
                        <span class="badge badge-secondary">DRAFT</span>
                      }
                    </td>
                    <td class="text-right">
                      <div class="flex justify-end gap-2">
                        <button 
                          (click)="openEditModal(p)" 
                          class="btn btn-outline btn-sm text-primary"
                          title="Edit Produk"
                        >
                          ✏️ Edit
                        </button>
                        <button 
                          (click)="deleteProduct(p.id, p.title)" 
                          class="btn btn-outline btn-sm text-danger"
                          title="Hapus Produk"
                        >
                          🗑️ Hapus
                        </button>
                      </div>
                    </td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        }
      </div>

      <!-- Add / Edit Product Modal -->
      @if (showAddModal()) {
        <div class="modal-backdrop">
          <div class="modal-card">
            <div class="modal-header">
              <h3>{{ isEditMode() ? 'Edit Data Produk' : 'Tambah Produk Baru ke Katalog' }}</h3>
              <button (click)="closeAddModal()" class="close-btn">&times;</button>
            </div>
            <div class="modal-body">
              <form (ngSubmit)="submitProduct()">
                <div class="grid grid-2 gap-4 mb-3">
                  <div class="form-group">
                    <label class="form-label">Kode SKU *</label>
                    <input 
                      type="text" 
                      [(ngModel)]="newProduct.sku" 
                      name="sku" 
                      class="form-control" 
                      placeholder="SKU-BRG-001" 
                      [disabled]="isEditMode()"
                      required
                    />
                  </div>
                  <div class="form-group">
                    <label class="form-label">Harga Dasar (Rp) *</label>
                    <input 
                      type="number" 
                      [(ngModel)]="newProduct.base_price" 
                      name="base_price" 
                      class="form-control" 
                      min="100" 
                      placeholder="25000" 
                      required
                    />
                  </div>
                </div>

                <div class="form-group mb-3">
                  <label class="form-label">Nama Produk *</label>
                  <input 
                    type="text" 
                    [(ngModel)]="newProduct.title" 
                    name="title" 
                    class="form-control" 
                    placeholder="Contoh: Beras Rojo Lele Super 5kg" 
                    required
                  />
                </div>

                <div class="grid grid-2 gap-4 mb-3">
                  <div class="form-group">
                    <label class="form-label">Kategori</label>
                    <select [(ngModel)]="newProduct.category_id" name="category_id" class="form-control">
                      <option value="">-- Pilih Kategori --</option>
                      @for (c of categories(); track c.id) {
                        <option [value]="c.id">{{ c.category_name }}</option>
                      }
                    </select>
                  </div>
                  <div class="form-group">
                    <label class="form-label">Merek (Brand)</label>
                    <select [(ngModel)]="newProduct.brand_id" name="brand_id" class="form-control">
                      <option value="">-- Pilih Merek --</option>
                      @for (b of brands(); track b.id) {
                        <option [value]="b.id">{{ b.brand_name }}</option>
                      }
                    </select>
                  </div>
                </div>

                <div class="grid grid-2 gap-4 mb-3">
                  <div class="form-group">
                    <label class="form-label">Berat Satuan (Gram) *</label>
                    <input 
                      type="number" 
                      [(ngModel)]="newProduct.weight_gram" 
                      name="weight_gram" 
                      class="form-control" 
                      placeholder="1000" 
                      required
                    />
                  </div>
                  <div class="form-group flex items-center pt-6">
                    <label class="flex items-center gap-2 cursor-pointer">
                      <input 
                        type="checkbox" 
                        [(ngModel)]="newProduct.is_published" 
                        name="is_published" 
                      />
                      <span class="font-semibold">Langsung Tayangkan (Publish)</span>
                    </label>
                  </div>
                </div>

                <div class="form-group mb-3">
                  <label class="form-label">URL Foto Produk</label>
                  <input 
                    type="text" 
                    [(ngModel)]="newProduct.image_url" 
                    name="image_url" 
                    class="form-control" 
                    placeholder="https://images.unsplash.com/..." 
                  />
                </div>

                <div class="form-group mb-4">
                  <label class="form-label">Deskripsi Produk</label>
                  <textarea 
                    [(ngModel)]="newProduct.description" 
                    name="description" 
                    rows="2" 
                    class="form-control" 
                    placeholder="Spesifikasi, nutrisi, atau keunggulan produk..."
                  ></textarea>
                </div>

                <div class="modal-footer">
                  <button type="button" (click)="closeAddModal()" class="btn btn-outline" [disabled]="submitting()">Batal</button>
                  <button type="submit" class="btn btn-primary" [disabled]="submitting() || !isFormValid()">
                    @if (submitting()) { Menyimpan... } @else { {{ isEditMode() ? 'Perbarui Produk' : 'Simpan Produk' }} }
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    .catalog-mgmt-page {
      padding-bottom: 3rem;
    }
    .page-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 2rem;
      gap: 1rem;
    }
    .page-title {
      font-size: 1.875rem;
      font-weight: 800;
      color: var(--text-primary);
      margin: 0.5rem 0 0.25rem;
    }
    .badge-purple {
      background: #f3e8ff;
      color: #7e22ce;
      font-weight: 700;
      font-size: 0.75rem;
      padding: 0.25rem 0.5rem;
      border-radius: var(--radius-sm);
    }
    .search-box {
      max-width: 320px;
      width: 100%;
    }
    .product-thumb {
      width: 44px;
      height: 44px;
      border-radius: var(--radius-md);
      object-fit: cover;
      border: 1px solid var(--border-color);
    }
    .product-info-cell {
      display: flex;
      flex-direction: column;
    }
    .modal-backdrop {
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.5);
      backdrop-filter: blur(4px);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 2000;
      padding: 1rem;
    }
    .modal-card {
      background: var(--bg-surface);
      border-radius: var(--radius-lg);
      width: 100%;
      max-width: 580px;
      box-shadow: var(--shadow-lg);
      border: 1px solid var(--border-color);
      overflow: hidden;
    }
    .modal-header {
      padding: 1.25rem 1.5rem;
      border-bottom: 1px solid var(--border-color);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    .close-btn {
      background: none;
      border: none;
      font-size: 1.5rem;
      cursor: pointer;
      color: var(--text-muted);
    }
    .modal-body {
      padding: 1.5rem;
    }
    .modal-footer {
      display: flex;
      justify-content: flex-end;
      gap: 0.75rem;
      margin-top: 1.5rem;
    }
  `]
})
export class CatalogMgmtComponent implements OnInit {
  private catalogService = inject(CatalogService);
  private notify = inject(NotificationService);

  products = signal<Product[]>([]);
  categories = signal<Category[]>([]);
  brands = signal<Brand[]>([]);

  loading = signal(true);
  submitting = signal(false);
  showAddModal = signal(false);
  isEditMode = signal(false);
  editingProductId = signal<string | null>(null);
  searchQuery = '';

  newProduct = {
    sku: '',
    title: '',
    description: '',
    base_price: 25000,
    category_id: '',
    brand_id: '',
    weight_gram: 1000,
    image_url: 'https://images.unsplash.com/photo-1542838132-92c53300491e?w=500&auto=format&fit=crop',
    is_published: true
  };

  ngOnInit() {
    this.loadProducts();
    this.loadCategories();
    this.loadBrands();
  }

  loadProducts() {
    this.loading.set(true);
    this.catalogService.getProducts({ limit: 100 })
      .subscribe({
        next: (res) => {
          this.products.set(res.data || []);
          this.loading.set(false);
        },
        error: () => {
          this.notify.error('Gagal memuat katalog produk');
          this.loading.set(false);
        }
      });
  }

  loadCategories() {
    this.catalogService.getCategories()
      .subscribe({
        next: (res) => {
          if (res.data) this.categories.set(res.data);
        }
      });
  }

  loadBrands() {
    this.catalogService.getBrands()
      .subscribe({
        next: (res) => {
          if (res.data) this.brands.set(res.data);
        }
      });
  }

  filteredProducts(): Product[] {
    if (!this.searchQuery.trim()) return this.products();
    const q = this.searchQuery.toLowerCase();
    return this.products().filter(p =>
      p.title.toLowerCase().includes(q) ||
      p.sku.toLowerCase().includes(q) ||
      (p.category_name && p.category_name.toLowerCase().includes(q))
    );
  }

  openAddModal() {
    this.isEditMode.set(false);
    this.editingProductId.set(null);
    this.newProduct = {
      sku: 'SKU-' + Math.floor(1000 + Math.random() * 9000),
      title: '',
      description: '',
      base_price: 20000,
      category_id: this.categories().length > 0 ? this.categories()[0].id : '',
      brand_id: this.brands().length > 0 ? this.brands()[0].id : '',
      weight_gram: 500,
      image_url: 'https://images.unsplash.com/photo-1542838132-92c53300491e?w=500&auto=format&fit=crop',
      is_published: true
    };
    this.showAddModal.set(true);
  }

  openEditModal(p: Product) {
    this.isEditMode.set(true);
    this.editingProductId.set(p.id);
    this.newProduct = {
      sku: p.sku,
      title: p.title,
      description: p.description || '',
      base_price: p.base_price,
      category_id: p.category_id || '',
      brand_id: p.brand_id || '',
      weight_gram: p.weight_gram,
      image_url: p.image_url || 'https://images.unsplash.com/photo-1542838132-92c53300491e?w=500&auto=format&fit=crop',
      is_published: p.is_published
    };
    this.showAddModal.set(true);
  }

  closeAddModal() {
    this.showAddModal.set(false);
  }

  isFormValid(): boolean {
    return !!this.newProduct.sku && !!this.newProduct.title && this.newProduct.base_price > 0;
  }

  submitProduct() {
    this.submitting.set(true);
    const payload = {
      title: this.newProduct.title,
      description: this.newProduct.description ? this.newProduct.description : null,
      base_price: Number(this.newProduct.base_price),
      category_id: this.newProduct.category_id || null,
      brand_id: this.newProduct.brand_id || null,
      weight_gram: Number(this.newProduct.weight_gram),
      image_url: this.newProduct.image_url ? this.newProduct.image_url : null,
      is_published: this.newProduct.is_published
    };

    if (this.isEditMode() && this.editingProductId()) {
      this.catalogService.updateProduct(this.editingProductId()!, payload).subscribe({
        next: () => {
          this.notify.success('Data produk berhasil diperbarui!');
          this.closeAddModal();
          this.submitting.set(false);
          this.loadProducts();
        },
        error: (err) => {
          this.notify.error(err.error?.message || 'Gagal memperbarui produk');
          this.submitting.set(false);
        }
      });
    } else {
      this.catalogService.createProduct({
        ...payload,
        sku: this.newProduct.sku
      }).subscribe({
        next: () => {
          this.notify.success('Produk baru berhasil ditambahkan ke katalog!');
          this.closeAddModal();
          this.submitting.set(false);
          this.loadProducts();
        },
        error: (err) => {
          this.notify.error(err.error?.message || 'Gagal menyimpan produk');
          this.submitting.set(false);
        }
      });
    }
  }

  deleteProduct(id: string, title: string) {
    if (!confirm(`Hapus produk "${title}" dari katalog?`)) return;

    this.catalogService.deleteProduct(id)
      .subscribe({
        next: () => {
          this.notify.success('Produk berhasil dihapus');
          this.loadProducts();
        },
        error: (err) => {
          this.notify.error(err.error?.message || 'Gagal menghapus produk');
        }
      });
  }
}
