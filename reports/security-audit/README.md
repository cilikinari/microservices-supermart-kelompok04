# Laporan Audit Keamanan Kontainer & Mitigasi Kerentanan CVE (Bagian IV)

**Mata Kuliah:** Arsitektur Mikroservis (TI-3142)  
**Tugas Terstruktur 4:** Rekayasa Kontainer Tingkat Lanjut: Optimasi Docker Image Footprint, Multi-Stage Builds, dan Hardening Keamanan Layanan Mikro  
**Layanan yang Dioptimasi:** Backend Go Service (`supermart-backend`)  
**Repositori:** `microservices-supermart-kelompok04`  
**Cabang Git:** `feat/docker-optimization`  

---

## 1. Ringkasan Eksekutif (Executive Summary)

Sebagai pemenuhan instruksi penugasan **BAGIAN IV: Pengerasan Keamanan (Hardening) & Audit Kerentanan (Bobot: 30 Poin)**, berkas ini memuat dokumentasi resmi serta tautan bukti log pemindaian keamanan empiris (*raw audit logs*) pada layanan **Backend Go** menggunakan kakas pemindai kerentanan standar industri **Trivy CLI**.

Audit dilakukan secara komparatif dengan membandingkan kondisi awal sebelum optimasi (*Baseline Single-Stage*) dengan kondisi akhir sesudah rekayasa ulang (*Multi-Stage Build Hardened*).

### Ringkasan Hasil Kunci Backend:
* **Penurunan Kerentanan Drastis:** Berhasil memangkas celah keamanan dari **158 CVE menjadi 5 CVE (reduksi 96.84%)**.
* **Eliminasi 100% Kerentanan Sistem Operasi:** Seluruh 153 celah kerentanan tingkat sistem operasi (OS-level) bawaan Debian berhasil dieliminasi total menjadi **0 CVE** melalui adopsi base image **Google Distroless Static Debian 12**.
* **Penegakan Prinsip Hak Akses Terendah (*Least Privilege*):** Kontainer produksi di-hardening untuk berjalan di bawah pengguna non-privilese (`USER nonroot:nonroot` dengan UID/GID 65532).

---

## 2. Matriks Evaluasi Penurunan Kerentanan (*Vulnerability Mitigation Matrix*)

Pemindaian dijalankan menggunakan filter tingkat keparahan tinggi sesuai panduan:  
`trivy image supermart-backend:[tag] --severity HIGH,CRITICAL`

| Target Docker Image | Tingkat Keparahan *Critical* | Tingkat Keparahan *High* | Tingkat Keparahan *Medium* | Tingkat Keparahan *Low* | Total Temuan Kerentanan |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Image Awal (Baseline Single-Stage)** | **3 CVE** | **155 CVE** | 0 CVE\* | 0 CVE\* | **158 CVE** |
| **Image Hasil Optimasi (Multi-Stage Distroless)** | **1 CVE** | **4 CVE** | 0 CVE\* | 0 CVE\* | **5 CVE** |
| **Selisih Penurunan Kerentanan (%)** | **-66.67%** | **-97.42%** | **0%** | **0%** | **-96.84%** |

*\*Catatan: Kolom Medium & Low bernilai 0 karena pemindaian difilter khusus `--severity HIGH,CRITICAL` sesuai ketentuan tugas.*

### Formulasi & Perhitungan Selisih Penurunan:
$$\text{Selisih Penurunan (\%)} = \frac{\text{Total Baseline} - \text{Total Distroless}}{\text{Total Baseline}} \times 100\%$$

* **Tingkat Critical:** $\frac{3 - 1}{3} \times 100\% = \mathbf{-66.67\%}$
* **Tingkat High:** $\frac{155 - 4}{155} \times 100\% = \mathbf{-97.42\%}$
* **Total Kerentanan:** $\frac{158 - 5}{158} \times 100\% = \mathbf{-96.84\%}$

---

## 3. Pembuktian Empiris Eksekusi Non-Root User (*Least Privilege*)

Secara *default*, kontainer Docker berjalan sebagai `root` (UID 0). Jika terjadi celah eksekusi kode arbitrer (*Remote Code Execution*), penyerang secara otomatis memperoleh hak akses administratif di dalam kontainer yang memperbesar peluang terjadinya eskalasi hak akses (*container escape / breakout*) ke host OS.

### A. Konfigurasi Dockerfile (`backend/Dockerfile`)
Layanan backend dikonfigurasi menggunakan pola Multi-Stage Build dengan base image **Google Distroless Static Debian 12**:
```dockerfile
# STAGE 2: Minimal Runtime Environment (Distroless Non-Root)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Salin zona waktu & sertifikat SSL dari builder stage
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Salin biner hasil kompilasi
COPY --from=builder /bin/supermart-backend /app/supermart-backend

# Terapkan Principle of Least Privilege (UID/GID 65532)
USER nonroot:nonroot

EXPOSE 3000

ENTRYPOINT ["/app/supermart-backend"]
```

### B. Bukti Verifikasi Empiris:
Karena image Distroless tidak memiliki *command interpreter / shell* (`/bin/sh` ditiadakan untuk pengerasan keamanan), verifikasi identitas pengguna dijalankan melalui inspeksi konfigurasi Docker daemon:

```bash
# 1. Jalankan kontainer backend
$ docker run -d --name test-backend supermart-backend:distroless

# 2. Periksa identitas pengguna yang menjalankan proses
$ docker inspect --format 'User: {{.Config.User}}' test-backend
User: nonroot:nonroot
```

*Hasil:* Kontainer terbukti secara empiris berjalan di bawah akun pengguna non-privilese `nonroot:nonroot` (UID/GID 65532).

---

## 4. Analisis Teknis & Evaluasi Keamanan

### A. Mengapa Base Image Google Distroless Lebih Aman?
1. **Zero Shell Environment:** Image Distroless tidak menyertakan *shell interpreter* (`/bin/sh` maupun `/bin/bash`). Serangan injeksi perintah (*Command Injection*) otomatis gagal dieksekusi oleh sistem operasi karena ketiadaan kakas penerjemah skrip.
2. **Zero Package Manager:** Meniadakan utilitas manajer paket seperti `apt` atau `dpkg`, sehingga penyerang tidak dapat mengunduh atau menginstal perangkat lunak berbahaya tambahan (*malware*, *netcat*, *curl*) di dalam kontainer yang sedang berjalan.
3. **Attack Surface Reduction:** Mengeliminasi seluruh pustaka sistem operasi Debian yang tidak esensial, memangkas celah kerentanan tingkat OS dari **153 CVE menjadi 0 CVE (eliminasi 100%)**.

### B. Pembedahan Kerentanan OS vs Kerentanan Dependensi Aplikasi (SCA)
Dari hasil pemindaian Trivy pada `supermart-backend:distroless`:
* **Kerentanan Tingkat OS (Debian):** **0 CVE** (Bersih total).
* **Kerentanan Tingkat Biner Go (`app/supermart-backend`):** **5 CVE** (1 Critical, 4 High).

**Analisis:**  
Kelima celah tersebut (`CVE-2026-25891`, `CVE-2026-25899`, `CVE-2025-66565`, `CVE-2026-63209`, dan `CVE-2026-78669`) berasal dari pustaka pihak ketiga bahasa Go yang tercantum pada berkas `backend/go.mod` (seperti framework `Fiber v3-beta` dan paket `klauspost/compress`). 

Hal ini membuktikan bahwa rekayasa kontainer berhasil mengamankan ranah infrastruktur/OS secara sempurna, sedangkan sisa 5 kerentanan yang ada berada pada ranah *Software Composition Analysis* (SCA) yang penyelesaiannya memerlukan pembaruan versi pustaka pada level kode sumber aplikasi.

---

## 5. Indeks Berkas Bukti Log Audit (*Audit Evidence*)

Berkas log mentah (*raw scan logs*) pada direktori ini dihasilkan langsung dari eksekusi perintah terminal berikut:

| Nama Berkas Log | Target Image | Deskripsi Berkas |
| :--- | :--- | :--- |
| [`backend-baseline-cve.txt`](./backend-baseline-cve.txt) | `supermart-backend:baseline` | Hasil scan image naive single-stage Go (`golang:latest`), memuat 158 CVE. |
| [`backend-distroless-cve.txt`](./backend-distroless-cve.txt) | `supermart-backend:distroless` | Hasil scan image multi-stage Distroless, memuat 5 CVE (0 CVE pada OS). |

### Perintah Reproduksi Audit (*Reproduction Commands*):
```bash
# 1. Pindai Image Baseline
trivy image supermart-backend:baseline --severity HIGH,CRITICAL > reports/security-audit/backend-baseline-cve.txt

# 2. Pindai Image Hasil Optimasi (Distroless)
trivy image supermart-backend:distroless --severity HIGH,CRITICAL > reports/security-audit/backend-distroless-cve.txt
```
