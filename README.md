# Backend URL Shortener

Aplikasi backend untuk layanan URL Shortener yang dibangun dengan Go (Golang) dan framework Gin. Aplikasi ini menyediakan API untuk memendekkan URL, melacak statistik klik, dan manajemen pengguna.

## 🚀 Fitur

- **Autentikasi Pengguna**
  - Registrasi dan login pengguna
  - JWT token untuk autentikasi
  - Ubah password pengguna

- **URL Shortener**
  - Buat URL pendek secara otomatis
  - Redirect dari URL pendek ke URL asli
  - Kelola URL (edit, hapus)
  - Generate short code unik

- **Analitik & Statistik**
  - Tracking jumlah klik per URL
  - Statistik detail per short code
  - Analytics keseluruhan
  - Record waktu klik

- **API RESTful**
  - CORS support untuk frontend
  - Middleware autentikasi
  - Response format JSON

## 📋 Prasyarat

- Go 1.25.6 atau lebih tinggi
- SQLite (sudah termasuk dalam dependencies)
- Git

## 🛠️ Instalasi & Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd backend-go
```

### 2. Install Dependencies
```bash
go mod download
```
atau
```bash
go install
```

### 3. Jalankan Aplikasi
```bash
go run main.go
```

Aplikasi akan berjalan pada port `3000` (http://localhost:3000)

### 4. Build untuk Production
```bash
go build -o backend-go.exe
```
Kemudian jalankan:
```bash
./backend-go.exe
```

## 📁 Struktur Database

Aplikasi menggunakan database SQLite dengan tabel berikut:

### Users
- `id` - Primary Key
- `name` - Nama lengkap pengguna
- `username` - Username unik
- `email` - Email unik
- `password` - Password (hashed)
- `created_at` - Waktu pembuatan
- `updated_at` - Waktu update

### URLs
- `id` - Primary Key
- `original_url` - URL asli
- `short_code` - Kode pendek unik
- `click_count` - Jumlah klik
- `user_id` - ID pemilik URL
- `created_at` - Waktu pembuatan
- `updated_at` - Waktu update

### Clicks
- `id` - Primary Key
- `url_id` - ID URL yang diklik
- `clicked_at` - Waktu klik
- `created_at` - Waktu pembuatan record
- `updated_at` - Waktu update record

## 🔗 API Endpoints

### Public Endpoints
- `GET /ping` - Health check
- `GET /:shortCode` - Redirect ke URL asli

### Authentication
- `POST /api/register` - Registrasi pengguna baru
- `POST /api/login` - Login pengguna

### Protected Endpoints (memerlukan authentication)
- `POST /api/shorten` - Buat URL pendek
- `GET /api/urls` - Dapatkan semua URL milik user
- `GET /api/stats/:shortCode` - Statistik per short code
- `PUT /api/urls/:id` - Update URL
- `DELETE /api/urls/:id` - Hapus URL
- `POST /api/change-password` - Ubah password
- `GET /api/analytics` - Analytics keseluruhan

## 📝 Contoh Penggunaan API

### Registrasi
```bash
curl -X POST http://localhost:3000/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "password123"
  }'
```

### Buat URL Pendek
```bash
curl -X POST http://localhost:3000/api/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "original_url": "https://example.com/very-long-url"
  }'
```

## 🔧 Konfigurasi

### CORS Configuration
Aplikasi dikonfigurasi untuk mengizinkan request dari:
- `http://localhost:3001`
- `http://127.0.0.1:3001`

### Database
- Menggunakan SQLite dengan nama file `db.sqlite`
- Auto-migration akan dijalankan saat aplikasi start
- Database akan dibuat otomatis jika belum ada

## 🛡️ Keamanan

- Password disimpan dalam bentuk hashed
- JWT token untuk autentikasi
- CORS configuration untuk keamanan cross-origin
- Input validation pada semua endpoint

## 📦 Dependencies

- `github.com/gin-gonic/gin` - Web framework
- `github.com/gin-contrib/cors` - CORS middleware
- `gorm.io/gorm` - ORM database
- `gorm.io/driver/sqlite` - SQLite driver
- `github.com/golang-jwt/jwt/v5` - JWT implementation
- `github.com/google/uuid` - UUID generation

## 🚨 Troubleshooting

### Port 3000 sudah digunakan
Ubah port di `main.go` pada baris terakhir:
```go
r.Run(":8080") // ganti dengan port yang tersedia
```

### Database error
Pastikan file `db.sqlite` memiliki permission yang tepat. Aplikasi akan otomatis membuat database jika belum ada.

### Dependency issues
Jalankan:
```bash
go mod tidy
```

## 🤝 Kontribusi

1. Fork repository
2. Buat branch baru (`git checkout -b feature/AmazingFeature`)
3. Commit perubahan (`git commit -m 'Add some AmazingFeature'`)
4. Push ke branch (`git push origin feature/AmazingFeature`)
5. Buka Pull Request

## 📄 License

Project ini dilisensikan under MIT License - lihat file LICENSE untuk detailnya.

## 📞 Contact

Jika ada pertanyaan atau masalah, silakan hubungi developer atau buat issue di repository.
