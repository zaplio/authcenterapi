package repository

import (
	"authcenterapi/internal/provider"
	model "authcenterapi/model/entity"
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// UserRepository interface for user-related operations
type UserRepository interface {
	// User CRUD operations
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateLoginInfo(ctx context.Context, userID uuid.UUID, loginIP net.IP) error
	IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error
	ResetFailedAttempts(ctx context.Context, userID uuid.UUID) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

// ProfileRepository interface for profile-related operations
type ProfileRepository interface {
	Create(ctx context.Context, profile *model.Profile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Profile, error)
	Update(ctx context.Context, profile *model.Profile) error
	GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*model.UserWithProfile, error)
}

// TokenRepository interface for token-related operations
type TokenRepository interface {
	Create(ctx context.Context, token *model.TokenAccess) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.TokenAccess, error)
	UpdateLastUsed(ctx context.Context, tokenID uuid.UUID, ip net.IP) error
	RevokeToken(ctx context.Context, tokenID uuid.UUID, revokedBy *uuid.UUID, reason string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, revokedBy *uuid.UUID, reason string) error
	GetTokenWithUser(ctx context.Context, tokenHash string) (*model.TokenAccess, *model.User, error)
}

// PasswordResetRepository interface for password reset operations
type PasswordResetRepository interface {
	Create(ctx context.Context, reset *model.PasswordReset) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordReset, error)
	MarkAsUsed(ctx context.Context, resetID uuid.UUID, usedIP net.IP, userAgent string) error
}

// PermissionRepository interface for permission operations
type PermissionRepository interface {
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error)
}

// Repository aggregates all repository interfaces
type IRepository struct {
	User          UserRepository
	Profile       ProfileRepository
	Token         TokenRepository
	PasswordReset PasswordResetRepository
	Permission    PermissionRepository
}

// repo is the base struct that implements common functionality
type repo struct {
	logger provider.ILogger
	conn   *pgx.Conn
}

// NewRepository creates a new repository instance with all sub-repositories
func NewRepository(logger provider.ILogger, conn *pgx.Conn) *IRepository {
	return &IRepository{
		User:          NewUserRepository(logger, conn),
		Profile:       NewProfileRepository(logger, conn),
		Token:         NewTokenRepository(logger, conn),
		PasswordReset: NewPasswordResetRepository(logger, conn),
		Permission:    NewPermissionRepository(logger, conn),
	}
}
