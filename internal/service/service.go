package service

import (
	"authcenterapi/internal/provider"
	"authcenterapi/internal/repository"
	model "authcenterapi/model/entity"
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type IService interface {
	// User Authentication
	Register(ctx context.Context, req *model.CreateUserRequest) (*model.UserWithProfile, error)
	Login(ctx context.Context, req *model.LoginRequest, clientIP net.IP, userAgent string) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.LoginResponse, error)
	Logout(ctx context.Context, tokenHash string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// Password Management
	ChangePassword(ctx context.Context, userID uuid.UUID, req *model.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *model.ResetPasswordRequest) error
	ResetPassword(ctx context.Context, req *model.ConfirmResetPasswordRequest, clientIP net.IP, userAgent string) error

	// Email Verification
	SendEmailVerification(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error

	// OAuth/Social Login
	LoginWithGoogle(ctx context.Context, googleToken string, clientIP net.IP, userAgent string) (*model.LoginResponse, error)

	// Token Management
	ValidateToken(ctx context.Context, tokenHash string, clientIP net.IP) (*model.User, error)
	RevokeToken(ctx context.Context, tokenHash string) error

	// User Management
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.UserWithProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateProfileRequest) (*model.Profile, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error)
}

type service struct {
	logger provider.ILogger
	redis  *redis.Client
	repo   *repository.IRepository
}

func NewService(logger provider.ILogger, redis *redis.Client, repo *repository.IRepository) IService {
	return &service{
		logger: logger,
		redis:  redis,
		repo:   repo,
	}
}
