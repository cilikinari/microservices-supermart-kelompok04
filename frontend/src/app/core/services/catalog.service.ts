import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Product, Category, Brand, ProductReview, ProductDetailData, ApiResponse } from '../models/types';

export interface ProductListParams {
  search?: string;
  category_id?: string;
  brand_id?: string;
  min_price?: number;
  max_price?: number;
  page?: number;
  limit?: number;
  sort_by?: string;
}

export interface CreateProductPayload {
  sku: string;
  title: string;
  description?: string | null;
  base_price: number;
  brand_id?: string | null;
  category_id?: string | null;
  weight_gram: number;
  image_url?: string | null;
  is_published: boolean;
}

export interface UpdateProductPayload {
  title?: string;
  description?: string | null;
  base_price?: number;
  brand_id?: string | null;
  category_id?: string | null;
  weight_gram?: number;
  image_url?: string | null;
  is_published?: boolean;
}

@Injectable({ providedIn: 'root' })
export class CatalogService {
  private http = inject(HttpClient);
  private readonly baseUrl = '/api/v1/catalog';

  getProducts(params?: ProductListParams): Observable<ApiResponse<Product[]>> {
    let httpParams = new HttpParams();
    if (params) {
      if (params.search) httpParams = httpParams.set('search', params.search);
      if (params.category_id) httpParams = httpParams.set('category_id', params.category_id);
      if (params.brand_id) httpParams = httpParams.set('brand_id', params.brand_id);
      if (params.min_price != null) httpParams = httpParams.set('min_price', params.min_price.toString());
      if (params.max_price != null) httpParams = httpParams.set('max_price', params.max_price.toString());
      if (params.page != null) httpParams = httpParams.set('page', params.page.toString());
      if (params.limit != null) httpParams = httpParams.set('limit', params.limit.toString());
      if (params.sort_by) httpParams = httpParams.set('sort_by', params.sort_by);
    }
    return this.http.get<ApiResponse<Product[]>>(`${this.baseUrl}/products`, { params: httpParams });
  }

  getProduct(id: string): Observable<ApiResponse<ProductDetailData>> {
    return this.http.get<ApiResponse<ProductDetailData>>(`${this.baseUrl}/products/${id}`);
  }

  getCategories(): Observable<ApiResponse<Category[]>> {
    return this.http.get<ApiResponse<Category[]>>(`${this.baseUrl}/categories`);
  }

  getBrands(): Observable<ApiResponse<Brand[]>> {
    return this.http.get<ApiResponse<Brand[]>>(`${this.baseUrl}/brands`);
  }

  getReviews(productId: string): Observable<ApiResponse<ProductReview[]>> {
    return this.http.get<ApiResponse<ProductReview[]>>(`${this.baseUrl}/products/${productId}/reviews`);
  }

  addReview(productId: string, rating: number, reviewText?: string): Observable<ApiResponse<any>> {
    return this.http.post<ApiResponse<any>>(`${this.baseUrl}/products/${productId}/reviews`, {
      rating,
      review_text: reviewText || null
    });
  }

  createProduct(payload: CreateProductPayload): Observable<ApiResponse<Product>> {
    return this.http.post<ApiResponse<Product>>(`${this.baseUrl}/products`, payload);
  }

  updateProduct(id: string, payload: UpdateProductPayload): Observable<ApiResponse<any>> {
    return this.http.put<ApiResponse<any>>(`${this.baseUrl}/products/${id}`, payload);
  }

  deleteProduct(id: string): Observable<ApiResponse<any>> {
    return this.http.delete<ApiResponse<any>>(`${this.baseUrl}/products/${id}`);
  }

  createCategory(payload: { category_name: string; slug: string; icon_url?: string }): Observable<ApiResponse<Category>> {
    return this.http.post<ApiResponse<Category>>(`${this.baseUrl}/categories`, payload);
  }

  createBrand(payload: { brand_name: string; logo_url?: string; website_url?: string }): Observable<ApiResponse<Brand>> {
    return this.http.post<ApiResponse<Brand>>(`${this.baseUrl}/brands`, payload);
  }
}

