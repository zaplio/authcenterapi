# Repository Package

Repository package menyediakan data access layer untuk semua authentication entities dalam aplikasi. Package ini mengimplementasikan repository pattern dengan interface yang jelas untuk setiap entity.

## 📁 Structure

```
internal/repository/
├── repository.go              # Main repository interfaces dan constructor
├── event_repo.go             # Event repository implementation
├── user_repo.go              # User repository implementation
├── profile_repo.go           # Profile repository implementation
├── token_repo.go             # Token access repository implementation
├── password_reset_repo.go    # Password reset repository implementation
├── permission_repo.go        # Permission repository implementation
└── README.md                 # This documentation
```

## 🎯 Repository Pattern

### Design Principles
- **Interface Segregation**: Setiap entity memiliki interface terpisah
- **Dependency Injection**: Dependencies diinjection melalui constructor
- **Error Handling**: Consistent error handling dan logging
- **Transaction Support**: Ready untuk transaction support
- **Connection Pooling**: Menggunakan pgx connection pool

### Repository Aggregation
```go
type Repository struct {
    Event        EventRepository
    User         UserRepository
    Profile      ProfileRepository
    Token        TokenRepository
    PasswordReset PasswordResetRepository
    Permission   PermissionRepository
}
```

## 📋 Repository Interfaces

### 1. **UserRepository**
User management operations:
```go
type UserRepository interface {
    // CRUD operations
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    
    // Authentication-specific operations
    UpdateLoginInfo(ctx context.Context, userID uuid.UUID, loginIP net.IP) error
    IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error
    ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error
    UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
    
    // Management operations
    SoftDelete(ctx context.Context, userID uuid.UUID) error
    GetActiveUsers(ctx context.Context, limit, offset int) ([]*model.User, error)
    SearchUsers(ctx context.Context, query string, limit, offset int) ([]*model.UserWithProfile, error)
}
```

### 2. **ProfileRepository**
Profile management operations:
```go
type ProfileRepository interface {
    Create(ctx context.Context, profile *model.Profile) error
    GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Profile, error)
    Update(ctx context.Context, profile *model.Profile) error
    Delete(ctx context.Context, userID uuid.UUID) error
    GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*model.UserWithProfile, error)
}
```

### 3. **TokenRepository**
Token management operations:
```go
type TokenRepository interface {
    Create(ctx context.Context, token *model.TokenAccess) error
    GetByTokenHash(ctx context.Context, tokenHash string) (*model.TokenAccess, error)
    GetActiveTokensByUser(ctx context.Context, userID uuid.UUID) ([]*model.TokenAccess, error)
    UpdateLastUsed(ctx context.Context, tokenID uuid.UUID, ip net.IP) error
    RevokeToken(ctx context.Context, tokenID uuid.UUID, revokedBy *uuid.UUID, reason string) error
    RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, revokedBy *uuid.UUID, reason string) error
    CleanExpiredTokens(ctx context.Context) error
    GetTokenWithUser(ctx context.Context, tokenHash string) (*model.TokenAccess, *model.User, error)
}
```

### 4. **PasswordResetRepository**
Password reset operations:
```go
type PasswordResetRepository interface {
    Create(ctx context.Context, reset *model.PasswordReset) error
    GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordReset, error)
    IncrementAttempts(ctx context.Context, resetID uuid.UUID) error
    MarkAsUsed(ctx context.Context, resetID uuid.UUID, usedIP net.IP, userAgent string) error
    CleanExpiredResets(ctx context.Context) error
    GetActiveResetsByEmail(ctx context.Context, email string) ([]*model.PasswordReset, error)
}
```

### 5. **PermissionRepository**
Permission management operations:
```go
type PermissionRepository interface {
    Create(ctx context.Context, permission *model.Permission) error
    GetByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
    GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error)
    CheckUserPermission(ctx context.Context, userID uuid.UUID, role, permission string) (bool, error)
    GetUsersByRole(ctx context.Context, role string) ([]*model.UserWithProfile, error)
    RevokePermission(ctx context.Context, permissionID uuid.UUID, revokedBy uuid.UUID) error
    RevokeUserPermissions(ctx context.Context, userID uuid.UUID, revokedBy uuid.UUID) error
    GetUserPermissionView(ctx context.Context, userID uuid.UUID) ([]*model.UserPermissionView, error)
    UpdatePermission(ctx context.Context, permission *model.Permission) error
}
```

## 🚀 Usage Examples

### Repository Initialization
```go
package main

import (
    "authcenterapi/internal/repository"
    "authcenterapi/internal/provider"
)

func main() {
    // Initialize dependencies
    logger := provider.NewLogger()
    conn := provider.NewPostgresConnection()
    
    // Create repository
    repo := repository.NewRepository(logger, conn)
    
    // Use specific repositories
    user, err := repo.User.GetByEmail(ctx, "user@example.com")
    if err != nil {
        logger.Error("Failed to get user", "error", err)
        return
    }
    
    // Get user with profile
    userWithProfile, err := repo.Profile.GetUserWithProfile(ctx, user.ID)
    if err != nil {
        logger.Error("Failed to get user with profile", "error", err)
        return
    }
}
```

### User Management
```go
// Create new user
user := &model.User{
    Username:     "john_doe",
    Email:        "john@example.com",
    PasswordHash: hashedPassword,
    Status:       model.UserStatusActive,
}

err := repo.User.Create(ctx, user)
if err != nil {
    return err
}

// Create profile
profile := &model.Profile{
    UserID:      user.ID,
    FirstName:   &[]string{"John"}[0],
    LastName:    &[]string{"Doe"}[0],
    DisplayName: &[]string{"John Doe"}[0],
}

err = repo.Profile.Create(ctx, profile)
if err != nil {
    return err
}
```

### Authentication Flow
```go
// Login validation
user, err := repo.User.GetByUsernameOrEmail(ctx, loginRequest.Username)
if err != nil {
    return err
}

if user == nil {
    return errors.New("user not found")
}

// Validate password here...

// Update login info on success
err = repo.User.UpdateLoginInfo(ctx, user.ID, clientIP)
if err != nil {
    logger.Error("Failed to update login info", "error", err)
    // Continue anyway
}

// Create access token
token := &model.TokenAccess{
    UserID:    user.ID,
    TokenHash: tokenHash,
    TokenType: model.TokenTypeAccess,
    ExpiresAt: time.Now().Add(1 * time.Hour),
    UserAgent: &userAgent,
}

err = repo.Token.Create(ctx, token)
if err != nil {
    return err
}
```

### Token Validation
```go
// Validate token
token, user, err := repo.Token.GetTokenWithUser(ctx, tokenHash)
if err != nil {
    return err
}

if token == nil || user == nil {
    return errors.New("invalid token")
}

if !token.IsValid() || !user.IsActive() {
    return errors.New("token or user is not valid")
}

// Update last used
err = repo.Token.UpdateLastUsed(ctx, token.ID, clientIP)
if err != nil {
    logger.Error("Failed to update token last used", "error", err)
    // Continue anyway
}
```

### Permission Checking
```go
// Check if user has permission
hasPermission, err := repo.Permission.CheckUserPermission(ctx, userID, "admin", "user:write")
if err != nil {
    return err
}

if !hasPermission {
    return errors.New("access denied")
}

// Get all user permissions
permissions, err := repo.Permission.GetUserPermissions(ctx, userID)
if err != nil {
    return err
}

// Grant new permission
permission := &model.Permission{
    UserID:     userID,
    Role:       "moderator",
    Permission: "user:read",
    GrantedBy:  &adminUserID,
}

err = repo.Permission.Create(ctx, permission)
if err != nil {
    return err
}
```

### Password Reset Flow
```go
// Create password reset request
reset := &model.PasswordReset{
    UserID:    user.ID,
    Email:     user.Email,
    TokenHash: resetTokenHash,
    ExpiresAt: time.Now().Add(1 * time.Hour),
}

err := repo.PasswordReset.Create(ctx, reset)
if err != nil {
    return err
}

// Validate reset token
reset, err = repo.PasswordReset.GetByTokenHash(ctx, tokenHash)
if err != nil {
    return err
}

if reset == nil || !reset.IsValid() {
    return errors.New("invalid or expired reset token")
}

// Mark as used after successful password change
err = repo.PasswordReset.MarkAsUsed(ctx, reset.ID, clientIP, userAgent)
if err != nil {
    logger.Error("Failed to mark reset as used", "error", err)
}
```

## 🔧 Error Handling

### Common Error Patterns
```go
// Check for not found
user, err := repo.User.GetByID(ctx, userID)
if err != nil {
    return err
}
if user == nil {
    return errors.New("user not found")
}

// Check for no rows affected
err = repo.User.Update(ctx, user)
if err != nil {
    if err == sql.ErrNoRows {
        return errors.New("user not found or not updated")
    }
    return err
}
```

### Logging
Semua repository methods include comprehensive logging:
- **Info**: Successful operations
- **Error**: Failed operations dengan context
- **Warn**: Operations yang tidak affect any rows
- **Debug**: Detailed operations (token usage, etc.)

## 🛡️ Security Considerations

### Data Protection
- **Password Hashes**: Never logged or exposed
- **Token Hashes**: Never exposed in logs
- **Sensitive Fields**: Properly handled dalam logging

### SQL Injection Prevention
- **Parameterized Queries**: Semua queries menggunakan parameters
- **Input Validation**: Validation di service layer
- **Prepared Statements**: Menggunakan pgx prepared statements

### Rate Limiting Support
- **Failed Login Tracking**: Built-in failed attempt tracking
- **Account Locking**: Automatic account locking mechanism
- **Reset Attempt Limiting**: Password reset attempt limiting

## 📊 Performance Optimization

### Database Optimizations
- **Proper Indexing**: Queries optimized dengan database indexes
- **Connection Pooling**: Menggunakan pgx connection pool
- **Query Optimization**: Efficient queries dengan minimal joins

### Pagination Support
```go
// Get paginated users
users, err := repo.User.GetActiveUsers(ctx, limit, offset)

// Search with pagination
searchResults, err := repo.User.SearchUsers(ctx, query, limit, offset)
```

### Batch Operations
```go
// Revoke all user tokens
err := repo.Token.RevokeAllUserTokens(ctx, userID, revokedBy, "user logout")

// Revoke all user permissions
err := repo.Permission.RevokeUserPermissions(ctx, userID, revokedBy)
```

## 🧹 Maintenance Operations

### Cleanup Operations
```go
// Clean expired tokens (can be run as cron job)
err := repo.Token.CleanExpiredTokens(ctx)

// Clean expired password resets
err := repo.PasswordReset.CleanExpiredResets(ctx)
```

### Monitoring Queries
Repository methods provide detailed logging untuk monitoring:
- Operation counts
- Performance metrics
- Error rates
- Security events

## 🔄 Transaction Support

Untuk operations yang memerlukan transactions:
```go
// Begin transaction
tx, err := conn.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx)

// Use transaction in repository operations
// (Repository methods dapat dimodifikasi untuk accept tx)

// Commit transaction
err = tx.Commit(ctx)
if err != nil {
    return err
}
```

## 📚 Best Practices

### Repository Usage
1. **Context Usage**: Selalu pass context untuk cancellation
2. **Error Handling**: Handle semua errors appropriately
3. **Logging**: Use provided logging untuk debugging
4. **Validation**: Validate inputs di service layer
5. **Transactions**: Use transactions untuk related operations

### Performance
1. **Pagination**: Always use pagination untuk list operations
2. **Indexing**: Ensure proper database indexing
3. **Connection Pooling**: Use connection pooling
4. **Query Optimization**: Optimize queries untuk performance

### Security
1. **Input Sanitization**: Sanitize inputs di service layer
2. **Authorization**: Check permissions di service layer
3. **Audit Logging**: Log security-related operations
4. **Data Protection**: Protect sensitive data dalam logs
