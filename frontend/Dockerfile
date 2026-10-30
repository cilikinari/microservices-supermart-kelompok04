# Stage 1: Build Angular SPA
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm ci --prefer-offline --no-audit

# Copy source code and build production bundle
COPY . .
RUN npm run build -- --configuration production

# Stage 2: Serve with Nginx
FROM nginx:1.25-alpine

# Hapus konfigurasi default Nginx
RUN rm -rf /etc/nginx/conf.d/*

# Salin konfigurasi kustom Nginx (mendukung SPA Client-side Routing)
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Salin artefak statis HTML/JS/CSS HANYA dari builder stage
COPY --from=builder /app/dist/frontend/browser /usr/share/nginx/html

# Hardening: Jalankan Nginx sebagai non-root jika memungkinkan atau amankan permission
RUN touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid /usr/share/nginx/html /var/cache/nginx /var/log/nginx

USER nginx

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]