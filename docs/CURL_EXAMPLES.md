# cURL Testing Examples

## Overview
Koleksi contoh cURL commands untuk testing Authentication Center API tanpa menggunakan Postman.

## Environment Variables
Untuk memudahkan testing, set environment variables berikut:

### Linux/macOS:
```bash
export BASE_URL="http://localhost:8080/api/v1"
export ACCESS_TOKEN=""
export REFRESH_TOKEN=""
```

### Windows PowerShell:
```powershell
$BASE_URL = "http://localhost:8080/api/v1"
$ACCESS_TOKEN = ""
$REFRESH_TOKEN = ""
```

## System Endpoints

### Health Check
```bash
curl -X GET "$BASE_URL/health"
```

### API Version
```bash
curl -X GET "$BASE_URL/version"
```

## Authentication Endpoints

### 1. Register
```bash
curl -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePassword123!",
    "phone": "+628123456789"
  }'
```

### 2. Login
```bash
# Save response to extract tokens
curl -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "SecurePassword123!"
  }' \
  -o login_response.json

# Extract tokens (Linux/macOS)
export ACCESS_TOKEN=$(cat login_response.json | jq -r '.data.tokens.access_token')
export REFRESH_TOKEN=$(cat login_response.json | jq -r '.data.tokens.refresh_token')

# Extract tokens (Windows PowerShell with jq)
$response = Get-Content login_response.json | ConvertFrom-Json
$ACCESS_TOKEN = $response.data.tokens.access_token
$REFRESH_TOKEN = $response.data.tokens.refresh_token
```

### 3. Validate Token
```bash
curl -X GET "$BASE_URL/auth/validate" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 4. Refresh Token
```bash
curl -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }"
```

### 5. Logout
```bash
curl -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 6. Logout All Sessions
```bash
curl -X POST "$BASE_URL/auth/logout-all" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Password Management

### 1. Change Password
```bash
curl -X PUT "$BASE_URL/auth/change-password" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "SecurePassword123!",
    "new_password": "NewSecurePassword123!",
    "confirm_password": "NewSecurePassword123!"
  }'
```

### 2. Forgot Password
```bash
curl -X POST "$BASE_URL/auth/forgot-password" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com"
  }'
```

### 3. Reset Password
```bash
curl -X POST "$BASE_URL/auth/reset-password" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "reset_token_from_email",
    "new_password": "NewSecurePassword123!"
  }'
```

## Email Verification

### 1. Send Email Verification
```bash
curl -X POST "$BASE_URL/auth/send-email-verification" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 2. Verify Email
```bash
curl -X POST "$BASE_URL/auth/verify-email" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "verification_token_from_email"
  }'
```

## OAuth

### Google Login
```bash
curl -X POST "$BASE_URL/auth/google" \
  -H "Content-Type: application/json" \
  -d '{
    "google_token": "google_oauth_token"
  }'
```

## Profile Management

### 1. Get Profile
```bash
curl -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 2. Update Profile
```bash
curl -X PUT "$BASE_URL/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

## Permissions

### Get User Permissions
```bash
curl -X GET "$BASE_URL/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Advanced Examples

### 1. Complete Authentication Flow Script

#### Linux/macOS:
```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"

echo "=== Authentication Center API Testing ==="

# 1. Health Check
echo "1. Health Check..."
curl -s "$BASE_URL/health" | jq '.'

# 2. Register
echo -e "\n2. Register..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePassword123!",
    "phone": "+628123456789"
  }' | jq '.'

# 3. Login
echo -e "\n3. Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "SecurePassword123!"
  }')

echo "$LOGIN_RESPONSE" | jq '.'

# Extract tokens
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.tokens.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.tokens.refresh_token')

# 4. Get Profile
echo -e "\n4. Get Profile..."
curl -s -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'

# 5. Validate Token
echo -e "\n5. Validate Token..."
curl -s -X GET "$BASE_URL/auth/validate" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'

# 6. Logout
echo -e "\n6. Logout..."
curl -s -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'

echo -e "\n=== Testing Complete ==="
```

#### Windows PowerShell:
```powershell
# complete_test.ps1

$BASE_URL = "http://localhost:8080/api/v1"

Write-Host "=== Authentication Center API Testing ===" -ForegroundColor Green

# 1. Health Check
Write-Host "1. Health Check..." -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$BASE_URL/health" -Method GET
$response | ConvertTo-Json -Depth 3

# 2. Register
Write-Host "`n2. Register..." -ForegroundColor Yellow
$registerBody = @{
    username = "testuser"
    email = "test@example.com"
    password = "SecurePassword123!"
    phone = "+628123456789"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method POST -Body $registerBody -ContentType "application/json"
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Register failed (user might already exist): $($_.Exception.Message)" -ForegroundColor Red
}

# 3. Login
Write-Host "`n3. Login..." -ForegroundColor Yellow
$loginBody = @{
    username = "testuser"
    password = "SecurePassword123!"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$loginResponse | ConvertTo-Json -Depth 3

# Extract tokens
$ACCESS_TOKEN = $loginResponse.data.tokens.access_token
$REFRESH_TOKEN = $loginResponse.data.tokens.refresh_token

# 4. Get Profile
Write-Host "`n4. Get Profile..." -ForegroundColor Yellow
$headers = @{ Authorization = "Bearer $ACCESS_TOKEN" }
$response = Invoke-RestMethod -Uri "$BASE_URL/profile" -Method GET -Headers $headers
$response | ConvertTo-Json -Depth 3

# 5. Validate Token
Write-Host "`n5. Validate Token..." -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$BASE_URL/auth/validate" -Method GET -Headers $headers
$response | ConvertTo-Json -Depth 3

# 6. Logout
Write-Host "`n6. Logout..." -ForegroundColor Yellow
$response = Invoke-RestMethod -Uri "$BASE_URL/auth/logout" -Method POST -Headers $headers
$response | ConvertTo-Json -Depth 3

Write-Host "`n=== Testing Complete ===" -ForegroundColor Green
```

### 2. Load Testing Script
```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
CONCURRENT_USERS=10
REQUESTS_PER_USER=5

echo "=== Load Testing ==="
echo "Concurrent Users: $CONCURRENT_USERS"
echo "Requests per User: $REQUESTS_PER_USER"

for i in $(seq 1 $CONCURRENT_USERS); do
  {
    for j in $(seq 1 $REQUESTS_PER_USER); do
      curl -s -w "User $i Request $j: %{http_code} - %{time_total}s\n" \
        -o /dev/null \
        "$BASE_URL/health"
    done
  } &
done

wait
echo "=== Load Testing Complete ==="
```

## Error Testing

### 1. Invalid Credentials
```bash
curl -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "invalid_user",
    "password": "wrong_password"
  }'
```

### 2. Missing Authorization
```bash
curl -X GET "$BASE_URL/profile"
```

### 3. Invalid Token
```bash
curl -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer invalid_token"
```

### 4. Malformed JSON
```bash
curl -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username": "test", invalid_json}'
```

## Response Examples

### Successful Login Response:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "uuid-here",
      "username": "john_doe",
      "email": "john@example.com",
      "is_email_verified": false,
      "created_at": "2024-01-15T10:30:00Z"
    },
    "tokens": {
      "access_token": "jwt_access_token_here",
      "refresh_token": "jwt_refresh_token_here",
      "expires_in": 3600
    }
  }
}
```

### Error Response:
```json
{
  "success": false,
  "message": "Invalid credentials",
  "error": "INVALID_CREDENTIALS",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Notes
- Semua timestamp menggunakan format ISO 8601 (UTC)
- Password minimal 8 karakter dengan kombinasi huruf, angka, dan simbol
- Token access memiliki TTL 1 jam, refresh token 7 hari
- Rate limiting mungkin diterapkan untuk mencegah abuse
- Gunakan HTTPS di production environment
