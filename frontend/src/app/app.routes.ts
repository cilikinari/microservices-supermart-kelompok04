import { Routes } from '@angular/router';
import { MainLayoutComponent } from './layouts/main-layout/main-layout.component';
import { authGuard, roleGuard, customerGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: '',
    component: MainLayoutComponent,
    children: [
      { path: '', redirectTo: 'catalog', pathMatch: 'full' },
      
      // Auth routes
      {
        path: 'auth/login',
        loadComponent: () => import('./features/auth/login/login.component').then(m => m.LoginComponent)
      },
      {
        path: 'auth/register',
        loadComponent: () => import('./features/auth/register/register.component').then(m => m.RegisterComponent)
      },

      // Customer routes (Khusus Pelanggan)
      {
        path: 'catalog',
        loadComponent: () => import('./features/customer/catalog/catalog.component').then(m => m.CatalogComponent)
      },
      {
        path: 'catalog/:id',
        loadComponent: () => import('./features/customer/catalog/product-detail/product-detail.component').then(m => m.ProductDetailComponent)
      },
      {
        path: 'cart',
        loadComponent: () => import('./features/customer/cart/cart.component').then(m => m.CartComponent),
        canActivate: [customerGuard]
      },
      {
        path: 'checkout',
        loadComponent: () => import('./features/customer/checkout/checkout.component').then(m => m.CheckoutComponent),
        canActivate: [authGuard, customerGuard]
      },
      {
        path: 'orders',
        loadComponent: () => import('./features/customer/orders/orders.component').then(m => m.OrdersComponent),
        canActivate: [authGuard, customerGuard]
      },
      {
        path: 'wallet',
        loadComponent: () => import('./features/customer/wallet/wallet.component').then(m => m.WalletComponent),
        canActivate: [authGuard, customerGuard]
      },
      {
        path: 'support',
        loadComponent: () => import('./features/customer/support/support.component').then(m => m.SupportComponent),
        canActivate: [authGuard, customerGuard]
      },

      // Warehouse Staff routes
      {
        path: 'warehouse/stocks',
        loadComponent: () => import('./features/warehouse/stocks/stocks.component').then(m => m.StocksComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['WAREHOUSE_STAFF', 'SUPER_ADMIN'] }
      },
      {
        path: 'warehouse/mutations',
        loadComponent: () => import('./features/warehouse/mutations/mutations.component').then(m => m.MutationsComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['WAREHOUSE_STAFF', 'SUPER_ADMIN'] }
      },
      {
        path: 'warehouse/procurement',
        loadComponent: () => import('./features/warehouse/procurement/procurement.component').then(m => m.ProcurementComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['WAREHOUSE_STAFF', 'SUPER_ADMIN'] }
      },

      // Courier routes
      {
        path: 'courier/deliveries',
        loadComponent: () => import('./features/courier/deliveries/deliveries.component').then(m => m.DeliveriesComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['COURIER', 'SUPER_ADMIN'] }
      },

      // CS Agent routes
      {
        path: 'support-agent/tickets',
        loadComponent: () => import('./features/support-agent/tickets/tickets.component').then(m => m.SupportAgentTicketsComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['CS_AGENT', 'SUPER_ADMIN'] }
      },

      // Super Admin routes
      {
        path: 'admin',
        loadComponent: () => import('./features/admin/dashboard/dashboard.component').then(m => m.DashboardComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['SUPER_ADMIN'] }
      },
      {
        path: 'admin/users',
        loadComponent: () => import('./features/admin/users/users.component').then(m => m.UsersComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['SUPER_ADMIN'] }
      },
      {
        path: 'admin/catalog',
        loadComponent: () => import('./features/admin/catalog-mgmt/catalog-mgmt.component').then(m => m.CatalogMgmtComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['SUPER_ADMIN'] }
      },
      {
        path: 'admin/promotions',
        loadComponent: () => import('./features/admin/promotions/promotions.component').then(m => m.PromotionsComponent),
        canActivate: [authGuard, roleGuard],
        data: { roles: ['SUPER_ADMIN'] }
      }
    ]
  },
  { path: '**', redirectTo: 'catalog' }
];
