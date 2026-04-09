# Documentation

Folder ini berisi dokumentasi lengkap untuk Authentication Center API.

## 📋 Daftar Dokumentasi

### 1. [API Documentation](API_DOCUMENTATION.md)
Dokumentasi resmi API yang mencakup:
- Semua endpoint yang tersedia
- Request/Response format
- Authentication requirements
- Error codes dan handling
- Parameter dan validation rules

### 2. [Postman Collection](postman_collection.json)
Koleksi Postman lengkap untuk testing API:
- Semua endpoint dengan contoh request
- Auto-script untuk token management
- Pre-configured request bodies
- Test scripts untuk automation

### 3. [Postman Environment Files](postman_environment_dev.json)
Environment files untuk Postman:
- **Development:** `postman_environment_dev.json`
- **Production:** `postman_environment_prod.json`

### 4. [Postman Testing Guide](POSTMAN_TESTING_GUIDE.md)
Panduan lengkap menggunakan Postman untuk testing:
- Import collection dan environment
- Testing flow dan scenarios
- Tips dan troubleshooting
- Automated testing dengan Newman

### 5. [cURL Examples](CURL_EXAMPLES.md)
Contoh lengkap cURL commands untuk testing:
- Semua endpoint dengan contoh request
- Shell scripts untuk automation
- Error testing scenarios
- Load testing examples

## 🚀 Quick Start

### Menggunakan Postman
1. Import `postman_collection.json`
2. Import `postman_environment_dev.json`
3. Ikuti panduan di `POSTMAN_TESTING_GUIDE.md`

### Menggunakan cURL
1. Buka `CURL_EXAMPLES.md`
2. Copy command yang dibutuhkan
3. Sesuaikan environment variables

### Testing Flow
1. **Health Check** - Pastikan server berjalan
2. **Register** - Buat akun baru
3. **Login** - Dapatkan access token
4. **Protected Endpoints** - Test dengan authentication

## 📖 Struktur API

### Base URL
- **Development:** `http://localhost:8080/api/v1`
- **Production:** `https://your-domain.com/api/v1`

### Endpoint Categories
- **Authentication:** `/auth/*`
- **Profile Management:** `/profile`
- **Permissions:** `/permissions`
- **System:** `/health`, `/version`

### Authentication
API menggunakan JWT Bearer token:
```
Authorization: Bearer <access_token>
```

## 🔧 Environment Setup

### Prerequisites
- Go 1.23.3+
- PostgreSQL 15+
- Redis (optional, untuk session management)

### Database Schema
Gunakan file `../database/schema.sql` untuk membuat tabel.

### Configuration
Sesuaikan `../config.yaml` dengan environment Anda.

## 📝 Testing Checklist

### ✅ Functional Testing
- [ ] User registration
- [ ] User login/logout
- [ ] Token refresh
- [ ] Password management
- [ ] Email verification
- [ ] Profile management
- [ ] OAuth integration

### ✅ Security Testing
- [ ] Authentication validation
- [ ] Authorization checks
- [ ] Input validation
- [ ] SQL injection prevention
- [ ] XSS protection
- [ ] Rate limiting

### ✅ Performance Testing
- [ ] Response times
- [ ] Concurrent users
- [ ] Database connections
- [ ] Memory usage
- [ ] Load testing

## 🐛 Troubleshooting

### Common Issues
1. **Server not responding:** Check if server is running on correct port
2. **Database connection error:** Verify PostgreSQL is running and credentials are correct
3. **Token expired:** Use refresh token or login again
4. **Validation errors:** Check request format and required fields

### Debug Mode
Set environment variable untuk debug:
```bash
export DEBUG=true
export LOG_LEVEL=debug
```

## 📞 Support

Jika menemukan masalah atau butuh bantuan:
1. Periksa dokumentasi yang relevan
2. Cek error logs di server
3. Gunakan testing tools untuk isolasi masalah
4. Hubungi development team

## 📄 License

Dokumentasi ini adalah bagian dari Authentication Center API project.

---

**Last Updated:** 2024-01-15
**Version:** 1.0.0
