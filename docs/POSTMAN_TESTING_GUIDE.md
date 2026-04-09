# Postman Testing Guide

## Overview
Panduan ini akan membantu Anda menggunakan Postman Collection untuk menguji Authentication Center API.

## Import Collection dan Environment

### 1. Import Collection
1. Buka Postman
2. Klik **Import** di pojok kiri atas
3. Pilih file `postman_collection.json`
4. Klik **Import**

### 2. Import Environment
1. Klik ikon gear (⚙️) di pojok kanan atas
2. Klik **Import**
3. Pilih file environment yang sesuai:
   - `postman_environment_dev.json` untuk development
   - `postman_environment_prod.json` untuk production
4. Klik **Import**
5. Pilih environment yang telah diimport

## Testing Flow

### 1. Health Check
```
GET {{base_url}}/health
```
Pastikan API server berjalan dengan baik.

### 2. User Registration
```
POST {{base_url}}/auth/register
```
**Body:**
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePassword123!",
  "phone": "+628123456789"
}
```

### 3. User Login
```
POST {{base_url}}/auth/login
```
**Body:**
```json
{
  "username": "testuser",
  "password": "SecurePassword123!"
}
```

**Auto Script:** Token akan otomatis disimpan ke environment variables.

### 4. Protected Endpoints
Setelah login berhasil, Anda dapat mengakses endpoint yang memerlukan authentication:

- **Get Profile:** `GET {{base_url}}/profile`
- **Update Profile:** `PUT {{base_url}}/profile`
- **Change Password:** `PUT {{base_url}}/auth/change-password`
- **Logout:** `POST {{base_url}}/auth/logout`

## Environment Variables

### Development Environment
- `base_url`: http://localhost:8080/api/v1
- `access_token`: (auto-filled after login)
- `refresh_token`: (auto-filled after login)
- `username`: john_doe
- `email`: john@example.com
- `password`: SecurePassword123!

### Production Environment
- `base_url`: https://your-domain.com/api/v1
- Semua token dan credential kosong untuk keamanan

## Test Scenarios

### Complete Authentication Flow
1. **Register** → Buat akun baru
2. **Login** → Dapatkan access & refresh token
3. **Validate Token** → Verifikasi token valid
4. **Get Profile** → Ambil data profil user
5. **Update Profile** → Update informasi profil
6. **Change Password** → Ganti password
7. **Refresh Token** → Perbarui access token
8. **Logout** → Keluar dari sesi

### Password Recovery Flow
1. **Forgot Password** → Kirim reset token ke email
2. **Reset Password** → Reset password dengan token

### Email Verification Flow
1. **Send Email Verification** → Kirim verification token
2. **Verify Email** → Verifikasi email dengan token

### OAuth Flow
1. **Google Login** → Login menggunakan Google OAuth token

## Tips Testing

### 1. Token Management
- Token secara otomatis disimpan setelah login berhasil
- Refresh token digunakan ketika access token expired
- Gunakan **Logout All Sessions** untuk membersihkan semua token

### 2. Error Testing
Test berbagai skenario error:
- Login dengan credential salah
- Akses endpoint tanpa token
- Gunakan token yang expired
- Submit data yang tidak valid

### 3. Performance Testing
- Test concurrent login dari multiple users
- Test rate limiting (jika diimplementasi)
- Monitor response time untuk setiap endpoint

### 4. Security Testing
- Test SQL injection pada input fields
- Test XSS pada text fields
- Verify CORS headers
- Check password strength validation

## Common Issues

### 1. Token Expired
**Error:** `401 Unauthorized`
**Solution:** Gunakan refresh token endpoint atau login ulang

### 2. Invalid Credentials
**Error:** `401 Invalid credentials`
**Solution:** Periksa username/email dan password

### 3. Validation Error
**Error:** `400 Bad Request`
**Solution:** Periksa format dan kelengkapan data input

### 4. Server Not Running
**Error:** Connection refused
**Solution:** Pastikan server berjalan di port yang benar

## Environment Setup

### Development
```bash
# Start PostgreSQL
docker run -d --name postgres-auth \
  -e POSTGRES_DB=authcenter \
  -e POSTGRES_USER=auth_user \
  -e POSTGRES_PASSWORD=auth_password \
  -p 5432:5432 \
  postgres:15

# Start Redis (optional, for session management)
docker run -d --name redis-auth \
  -p 6379:6379 \
  redis:7-alpine

# Run the application
go run cmd/main.go
```

### Production
Pastikan semua environment variables production telah di-set dengan benar sebelum testing.

## Automated Testing

Anda juga dapat menjalankan collection ini secara otomatis menggunakan Newman:

```bash
# Install Newman
npm install -g newman

# Run collection
newman run postman_collection.json \
  -e postman_environment_dev.json \
  --reporters cli,html \
  --reporter-html-export results.html
```

## Contact
Jika ada pertanyaan atau issues, silakan hubungi development team.
