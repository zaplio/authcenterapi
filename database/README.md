# Authentication Database Schema

Sistem database authentication lengkap untuk aplikasi Go dengan PostgreSQL yang mencakup manajemen user, profile, token, password reset, dan permissions.

## 📁 Structure

```
database/
├── schema.sql                           # Complete database schema
├── queries.sql                          # Common SQL queries
└── migrations/
    └── 001_create_authentication_schema.sql  # Migration file

model/
└── auth_models.go                       # Go models untuk database tables
```

## 🗂️ Schema Overview

### 1. **authentication.users**
Table utama untuk data authentication user:
- **Primary Key**: `id` (UUID)
- **Unique Fields**: `username`, `email`, `phone`
- **Security Fields**: `password_hash`, `two_factor_secret`, `failed_login_attempts`, `locked_until`
- **Status Tracking**: `status`, `last_login_at`, `email_verified_at`, `phone_verified_at`
- **Soft Delete**: `deleted_at`

### 2. **authentication.profiles**
Extended profile information untuk users:
- **Relation**: `user_id` → `users.id` (ONE-TO-ONE)
- **Personal Info**: `first_name`, `last_name`, `display_name`, `bio`, `date_of_birth`, `gender`
- **Location**: `country`, `state`, `city`, `address`, `postal_code`
- **Preferences**: `timezone`, `language`, `preferences` (JSONB)
- **Social**: `website_url`, `social_links` (JSONB)
- **Flexible Storage**: `metadata` (JSONB)

### 3. **authentication.token_access**
Manajemen access tokens, refresh tokens, dan API keys:
- **Token Types**: `access`, `refresh`, `api_key`, `verification`
- **Security**: `token_hash` (never store plain tokens), `scope`, `expires_at`
- **Tracking**: `last_used_at`, `last_used_ip`, `user_agent`, `device_info`
- **Revocation**: `revoked`, `revoked_at`, `revoked_by`, `revoke_reason`

### 4. **authentication.password_resets**
Password reset token management:
- **Security**: `token_hash`, `expires_at`, `attempts`, `max_attempts`
- **Rate Limiting**: `blocked_until` (untuk mencegah brute force)
- **Usage Tracking**: `used`, `used_at`, `used_ip`

### 5. **authentication.permissions**
Role-based access control (RBAC):
- **RBAC Fields**: `role`, `permission`, `resource`, `resource_id`
- **Time-based**: `expires_at`, `granted_at`
- **Audit Trail**: `granted_by`, `revoked_by`, `revoked_at`
- **Flexible Conditions**: `conditions` (JSONB untuk conditional permissions)

## 🚀 Installation & Setup

### 1. Run Migration
```bash
# Connect to PostgreSQL
psql -h your_host -U your_username -d your_database

# Run migration
\i database/migrations/001_create_authentication_schema.sql
```

### 2. Verify Installation
```sql
-- Check if schema and tables are created
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'authentication';

-- Check views
SELECT table_name 
FROM information_schema.views 
WHERE table_schema = 'authentication';
```

## 📖 Common Usage Examples

### User Registration
```sql
-- Create user
INSERT INTO authentication.users (username, email, password_hash) 
VALUES ('john_doe', 'john@example.com', '$2a$10$hashed_password');

-- Create profile
INSERT INTO authentication.profiles (user_id, first_name, last_name, display_name)
VALUES (
    (SELECT id FROM authentication.users WHERE username = 'john_doe'),
    'John', 'Doe', 'John Doe'
);
```

### User Login Validation
```sql
-- Validate credentials
SELECT id, username, email, password_hash, status, failed_login_attempts, locked_until
FROM authentication.users 
WHERE (email = $1 OR username = $1) 
  AND status = 'active' 
  AND deleted_at IS NULL;

-- Update login info after success
UPDATE authentication.users 
SET last_login_at = CURRENT_TIMESTAMP, 
    last_login_ip = $2, 
    failed_login_attempts = 0
WHERE id = $1;
```

### Token Management
```sql
-- Create access token
INSERT INTO authentication.token_access 
(user_id, token_hash, token_type, scope, expires_at)
VALUES ($1, $2, 'access', 'read,write', CURRENT_TIMESTAMP + INTERVAL '1 hour');

-- Validate token
SELECT t.*, u.username, u.status
FROM authentication.token_access t
JOIN authentication.users u ON t.user_id = u.id
WHERE t.token_hash = $1 
  AND t.revoked = FALSE 
  AND t.expires_at > CURRENT_TIMESTAMP
  AND u.status = 'active';
```

### Permission Checking
```sql
-- Check if user has permission
SELECT COUNT(*) > 0 as has_permission
FROM authentication.permissions p
JOIN authentication.users u ON p.user_id = u.id
WHERE p.user_id = $1 
  AND (p.role = $2 OR p.permission = $3)
  AND p.revoked = FALSE
  AND (p.expires_at IS NULL OR p.expires_at > CURRENT_TIMESTAMP)
  AND u.status = 'active';
```

## 🔧 Go Integration

### Using the Models
```go
package main

import (
    "authcenterapi/model"
    "time"
    "github.com/google/uuid"
)

func main() {
    // Create new user
    user := &model.User{
        ID:       uuid.New(),
        Username: "john_doe",
        Email:    "john@example.com",
        Status:   model.UserStatusActive,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Create profile
    profile := &model.Profile{
        ID:          uuid.New(),
        UserID:      user.ID,
        FirstName:   &[]string{"John"}[0],
        LastName:    &[]string{"Doe"}[0],
        DisplayName: &[]string{"John Doe"}[0],
        Timezone:    "UTC",
        Language:    "en",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // Check if user is active
    if user.IsActive() {
        // User is active and not deleted
    }

    // Check if user is locked
    if user.IsLocked() {
        // Account is temporarily locked
    }
}
```

### Login Request Example
```go
type LoginHandler struct {
    // dependencies
}

func (h *LoginHandler) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
    // Validate credentials
    user, err := h.userRepo.GetByUsernameOrEmail(req.Username)
    if err != nil {
        return nil, err
    }

    // Check password, account status, etc.
    if !user.IsActive() {
        return nil, errors.New("account is not active")
    }

    if user.IsLocked() {
        return nil, errors.New("account is temporarily locked")
    }

    // Create tokens
    accessToken, err := h.tokenService.CreateAccessToken(user.ID)
    if err != nil {
        return nil, err
    }

    refreshToken, err := h.tokenService.CreateRefreshToken(user.ID)
    if err != nil {
        return nil, err
    }

    return &model.LoginResponse{
        User:         model.UserWithProfile{User: *user, Profile: profile},
        AccessToken:  accessToken.Token,
        RefreshToken: refreshToken.Token,
        ExpiresAt:    accessToken.ExpiresAt,
    }, nil
}
```

## 🔒 Security Features

### 1. **Password Security**
- Passwords disimpan sebagai hash (recommend bcrypt atau Argon2)
- Never expose `password_hash` di JSON response
- Support untuk 2FA dengan `two_factor_secret`

### 2. **Account Security**
- Failed login attempt tracking dengan auto-lock mechanism
- Email dan phone verification tracking
- Soft delete support dengan `deleted_at`

### 3. **Token Security**
- Token disimpan sebagai hash, bukan plain text
- Support untuk token expiration dan revocation
- Device tracking dengan `user_agent` dan `device_info`
- Token scope untuk granular permissions

### 4. **Rate Limiting**
- Password reset dengan attempt limiting dan blocking
- Failed login attempts dengan account locking

### 5. **Audit Trail**
- Complete audit trail untuk permissions
- Tracking untuk semua login dan token usage
- Timestamping untuk semua operations

## 📊 Indexes dan Performance

### Optimized Indexes
- Primary keys dengan UUID v4
- Composite indexes untuk common query patterns
- Indexes untuk foreign keys dan frequently queried fields

### Query Optimization
- Views untuk common complex queries
- Proper use of JSONB untuk flexible data storage
- Efficient pagination dengan proper indexing

## 🔄 Views

### 1. **v_active_users_with_profiles**
Menggabungkan users dan profiles untuk active users:
```sql
SELECT * FROM authentication.v_active_users_with_profiles 
WHERE country = 'ID' 
ORDER BY user_created_at DESC;
```

### 2. **v_user_permissions**
View untuk melihat effective permissions user:
```sql
SELECT * FROM authentication.v_user_permissions 
WHERE user_id = $1;
```

## 🧹 Maintenance

### Regular Cleanup
```sql
-- Clean expired tokens (older than 30 days)
DELETE FROM authentication.token_access 
WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '30 days';

-- Clean old password reset tokens (older than 7 days)
DELETE FROM authentication.password_resets 
WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '7 days';
```

### Monitoring Queries
```sql
-- User statistics
SELECT 
    COUNT(*) as total_users,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
    COUNT(CASE WHEN two_factor_enabled = TRUE THEN 1 END) as two_factor_users
FROM authentication.users 
WHERE deleted_at IS NULL;

-- Login activity (last 7 days)
SELECT 
    DATE(last_login_at) as login_date,
    COUNT(DISTINCT id) as unique_logins
FROM authentication.users 
WHERE last_login_at > CURRENT_TIMESTAMP - INTERVAL '7 days'
GROUP BY DATE(last_login_at)
ORDER BY login_date DESC;
```

## 🎯 Best Practices

### 1. **Security**
- Always hash passwords dengan salt yang strong
- Use secure random untuk token generation
- Implement proper rate limiting
- Regular security audits

### 2. **Performance**
- Use connection pooling
- Implement proper caching strategy
- Regular VACUUM dan ANALYZE tables
- Monitor query performance

### 3. **Data Integrity**
- Use transactions untuk related operations
- Implement proper foreign key constraints
- Regular backup dan recovery testing

### 4. **Scalability**
- Consider partitioning untuk large tables
- Implement read replicas untuk read-heavy operations
- Use proper indexes untuk query optimization

## 📝 Configuration

Update `config.yaml` sesuai dengan database configuration:

```yaml
postgres:
  host: your_postgres_host
  port: 5432
  username: your_username
  password: your_password
  database: your_database
  options:
    - sslmode=require  # untuk production
```

## 🤝 Contributing

1. Pastikan schema changes kompatible dengan existing data
2. Tambahkan migration files untuk schema changes
3. Update model definitions sesuai dengan schema changes
4. Test semua queries dan ensure performance

## 📚 References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [UUID Best Practices](https://www.postgresql.org/docs/current/datatype-uuid.html)
- [JSONB in PostgreSQL](https://www.postgresql.org/docs/current/datatype-json.html)
- [Index Types](https://www.postgresql.org/docs/current/indexes-types.html)
