# Authentication Center API Documentation

## Overview
Authentication Center API adalah layanan autentikasi yang menyediakan fitur lengkap untuk manajemen user, autentikasi, dan otorisasi.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
API menggunakan Bearer Token authentication. Sertakan token di header:
```
Authorization: Bearer <your_access_token>
```

## Response Format
Semua response menggunakan format JSON dengan struktur standar:

### Success Response
```json
{
  "success": true,
  "message": "Success message",
  "data": {
    // Response data
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message",
  "error": {
    "code": "ERROR_CODE",
    "details": "Detailed error information"
  }
}
```

## Endpoints

### 1. User Registration

**POST** `/auth/register`

Mendaftarkan user baru ke dalam sistem.

#### Request Body
```json
{
  "username": "string",
  "email": "string",
  "password": "string",
  "phone": "string"
}
```

#### Request Example
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePassword123!",
  "phone": "+628123456789"
}
```

#### Response
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "phone": "+628123456789",
      "status": "active",
      "email_verified_at": null,
      "phone_verified_at": null,
      "two_factor_enabled": false,
      "created_at": "2025-08-27T10:00:00Z",
      "updated_at": "2025-08-27T10:00:00Z"
    },
    "profile": {
      "id": "uuid",
      "user_id": "uuid",
      "timezone": "UTC",
      "language": "en",
      "created_at": "2025-08-27T10:00:00Z",
      "updated_at": "2025-08-27T10:00:00Z"
    }
  }
}
```

#### Error Responses
- **400 Bad Request**: Invalid input data
- **409 Conflict**: Username atau email sudah terdaftar

---

### 2. User Login

**POST** `/auth/login`

Autentikasi user dan mendapatkan access token.

#### Request Body
```json
{
  "username": "string",
  "password": "string"
}
```

#### Request Example
```json
{
  "username": "john_doe",
  "password": "SecurePassword123!"
}
```

#### Response
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "status": "active",
      "last_login_at": "2025-08-27T10:00:00Z"
    },
    "profile": {
      "id": "uuid",
      "user_id": "uuid",
      "timezone": "UTC",
      "language": "en"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 3600
    }
  }
}
```

#### Error Responses
- **400 Bad Request**: Invalid input data
- **401 Unauthorized**: Username atau password salah
- **423 Locked**: Account terkunci karena terlalu banyak percobaan login

---

### 3. Refresh Token

**POST** `/auth/refresh`

Memperbarui access token menggunakan refresh token.

#### Request Body
```json
{
  "refresh_token": "string"
}
```

#### Response
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com"
    },
    "profile": {
      "id": "uuid",
      "user_id": "uuid"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 3600
    }
  }
}
```

#### Error Responses
- **400 Bad Request**: Invalid refresh token
- **401 Unauthorized**: Refresh token expired atau tidak valid

---

### 4. Logout

**POST** `/auth/logout`

Logout user dan revoke current token.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Response
```json
{
  "success": true,
  "message": "Logout successful"
}
```

#### Error Responses
- **401 Unauthorized**: Token tidak valid

---

### 5. Logout All Sessions

**POST** `/auth/logout-all`

Logout dari semua device/session.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Response
```json
{
  "success": true,
  "message": "Logged out from all sessions"
}
```

#### Error Responses
- **401 Unauthorized**: Token tidak valid

---

### 6. Change Password

**PUT** `/auth/change-password`

Mengubah password user.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Request Body
```json
{
  "current_password": "string",
  "new_password": "string",
  "confirm_password": "string"
}
```

#### Request Example
```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword123!",
  "confirm_password": "NewPassword123!"
}
```

#### Response
```json
{
  "success": true,
  "message": "Password changed successfully"
}
```

#### Error Responses
- **400 Bad Request**: Password confirmation tidak cocok
- **401 Unauthorized**: Current password salah

---

### 7. Forgot Password

**POST** `/auth/forgot-password`

Memulai proses reset password.

#### Request Body
```json
{
  "email": "string"
}
```

#### Request Example
```json
{
  "email": "john@example.com"
}
```

#### Response
```json
{
  "success": true,
  "message": "Password reset email sent"
}
```

#### Error Responses
- **404 Not Found**: Email tidak terdaftar

---

### 8. Reset Password

**POST** `/auth/reset-password`

Menyelesaikan proses reset password.

#### Request Body
```json
{
  "token": "string",
  "new_password": "string"
}
```

#### Request Example
```json
{
  "token": "reset_token_here",
  "new_password": "NewPassword123!"
}
```

#### Response
```json
{
  "success": true,
  "message": "Password reset successfully"
}
```

#### Error Responses
- **400 Bad Request**: Token tidak valid atau expired
- **404 Not Found**: Token tidak ditemukan

---

### 9. Send Email Verification

**POST** `/auth/send-email-verification`

Mengirim ulang email verifikasi.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Response
```json
{
  "success": true,
  "message": "Verification email sent"
}
```

#### Error Responses
- **401 Unauthorized**: Token tidak valid
- **409 Conflict**: Email sudah terverifikasi

---

### 10. Verify Email

**POST** `/auth/verify-email`

Memverifikasi email dengan token.

#### Request Body
```json
{
  "token": "string"
}
```

#### Response
```json
{
  "success": true,
  "message": "Email verified successfully"
}
```

#### Error Responses
- **400 Bad Request**: Token tidak valid atau expired

---

### 11. Google OAuth Login

**POST** `/auth/google`

Login menggunakan Google OAuth.

#### Request Body
```json
{
  "google_token": "string"
}
```

#### Response
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 3600
    }
  }
}
```

#### Error Responses
- **400 Bad Request**: Invalid Google token
- **401 Unauthorized**: Google token tidak valid

---

### 12. Validate Token

**GET** `/auth/validate`

Memvalidasi access token dan mendapatkan informasi user.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Response
```json
{
  "success": true,
  "message": "Token is valid",
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "status": "active",
      "email_verified_at": "2025-08-27T10:00:00Z",
      "two_factor_enabled": false,
      "last_login_at": "2025-08-27T10:00:00Z"
    }
  }
}
```

#### Error Responses
- **401 Unauthorized**: Token tidak valid atau expired

---

### 13. Get User Profile

**GET** `/profile`

Mendapatkan profil user yang sedang login.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Response
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "phone": "+628123456789",
      "status": "active",
      "email_verified_at": "2025-08-27T10:00:00Z",
      "phone_verified_at": null,
      "two_factor_enabled": false,
      "created_at": "2025-08-27T10:00:00Z",
      "updated_at": "2025-08-27T10:00:00Z"
    },
    "profile": {
      "id": "uuid",
      "user_id": "uuid",
      "first_name": "John",
      "last_name": "Doe",
      "display_name": "John Doe",
      "date_of_birth": "1990-01-01",
      "gender": "male",
      "country": "ID",
      "city": "Jakarta",
      "timezone": "Asia/Jakarta",
      "language": "id",
      "avatar_url": "https://example.com/avatar.jpg",
      "bio": "Software Developer",
      "created_at": "2025-08-27T10:00:00Z",
      "updated_at": "2025-08-27T10:00:00Z"
    }
  }
}
```

#### Error Responses
- **401 Unauthorized**: Token tidak valid
- **404 Not Found**: Profile tidak ditemukan

---

### 14. Update Profile

**PUT** `/profile`

Memperbarui profil user.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Request Body
```json
{
  "first_name": "string",
  "last_name": "string",
  "display_name": "string",
  "bio": "string",
  "date_of_birth": "string",
  "gender": "string",
  "country": "string",
  "state": "string",
  "city": "string",
  "address": "string",
  "postal_code": "string",
  "timezone": "string",
  "language": "string",
  "website_url": "string"
}
```

#### Request Example
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "display_name": "John Doe",
  "bio": "Full Stack Developer",
  "date_of_birth": "1990-01-01",
  "gender": "male",
  "country": "ID",
  "city": "Jakarta",
  "timezone": "Asia/Jakarta",
  "language": "id"
}
```

#### Response
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "first_name": "John",
    "last_name": "Doe",
    "display_name": "John Doe",
    "bio": "Full Stack Developer",
    "date_of_birth": "1990-01-01",
    "gender": "male",
    "country": "ID",
    "city": "Jakarta",
    "timezone": "Asia/Jakarta",
    "language": "id",
    "updated_at": "2025-08-27T10:00:00Z"
  }
}
```

#### Error Responses
- **400 Bad Request**: Invalid input data
- **401 Unauthorized**: Token tidak valid

---

### 15. Get User Permissions

**GET** `/permissions`

Mendapatkan daftar permissions user.

#### Headers
```
Authorization: Bearer <access_token>
```

#### Response
```json
{
  "success": true,
  "message": "Permissions retrieved successfully",
  "data": {
    "permissions": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "role": "admin",
        "permission": "users.read",
        "resource": "users",
        "resource_id": null,
        "granted_by": "uuid",
        "granted_at": "2025-08-27T10:00:00Z",
        "expires_at": null,
        "revoked": false,
        "created_at": "2025-08-27T10:00:00Z"
      }
    ]
  }
}
```

#### Error Responses
- **401 Unauthorized**: Token tidak valid

---

## Status Codes

| Code | Description |
|------|-------------|
| 200  | OK - Request berhasil |
| 201  | Created - Resource berhasil dibuat |
| 400  | Bad Request - Input tidak valid |
| 401  | Unauthorized - Authentication required |
| 403  | Forbidden - Tidak memiliki permission |
| 404  | Not Found - Resource tidak ditemukan |
| 409  | Conflict - Resource sudah ada |
| 422  | Unprocessable Entity - Validation error |
| 423  | Locked - Account terkunci |
| 500  | Internal Server Error - Server error |

## Error Codes

| Code | Description |
|------|-------------|
| `INVALID_CREDENTIALS` | Username atau password salah |
| `ACCOUNT_LOCKED` | Account terkunci |
| `EMAIL_NOT_VERIFIED` | Email belum diverifikasi |
| `TOKEN_EXPIRED` | Token sudah expired |
| `TOKEN_INVALID` | Token tidak valid |
| `USER_NOT_FOUND` | User tidak ditemukan |
| `EMAIL_ALREADY_EXISTS` | Email sudah terdaftar |
| `USERNAME_ALREADY_EXISTS` | Username sudah terdaftar |
| `INVALID_PASSWORD` | Password tidak memenuhi kriteria |
| `PASSWORD_MISMATCH` | Password confirmation tidak cocok |
| `INSUFFICIENT_PERMISSIONS` | Tidak memiliki permission yang cukup |

## Rate Limiting

API menggunakan rate limiting untuk mencegah abuse:

- **Login**: 5 requests per minute per IP
- **Registration**: 3 requests per minute per IP
- **Password Reset**: 3 requests per hour per email
- **General API**: 100 requests per minute per user

## Security Headers

API menyertakan security headers berikut:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

## CORS

API mendukung CORS dengan konfigurasi:

- **Allowed Origins**: Konfigurasi berdasarkan environment
- **Allowed Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Allowed Headers**: Authorization, Content-Type, X-Requested-With
- **Credentials**: true

## Environment Variables

Untuk development dan production, pastikan environment variables berikut dikonfigurasi:

```env
# Server
PORT=8080
HOST=localhost

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=authcenter
DB_USER=postgres
DB_PASSWORD=password

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=1h

# Email (untuk production)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Redis (optional untuk caching)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
```

## SDKs dan Examples

### cURL Examples

#### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePassword123!",
    "phone": "+628123456789"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "SecurePassword123!"
  }'
```

#### Get Profile
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer your-access-token"
```

### JavaScript Example
```javascript
// Login
const loginResponse = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    username: 'john_doe',
    password: 'SecurePassword123!'
  })
});

const loginData = await loginResponse.json();
const accessToken = loginData.data.tokens.access_token;

// Get Profile
const profileResponse = await fetch('/api/v1/profile', {
  headers: {
    'Authorization': `Bearer ${accessToken}`
  }
});

const profileData = await profileResponse.json();
```

## Testing

API menyediakan test endpoints untuk development:

**GET** `/health` - Health check
**GET** `/version` - API version info

## Support

Untuk pertanyaan atau dukungan teknis, silakan hubungi:
- Email: support@example.com
- Documentation: https://docs.example.com
- GitHub Issues: https://github.com/karatondev/authcenterapi/issues
