package service

import (
	"authcenterapi/internal/provider"
	model "authcenterapi/model/entity"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Token-related constants
const (
	AccessTokenExpiry       = 1 * time.Hour
	RefreshTokenExpiry      = 24 * time.Hour * 7 // 7 days
	VerificationTokenExpiry = 24 * time.Hour     // 24 hours
	ResetTokenExpiry        = 1 * time.Hour
)

// Register creates a new user account
func (s *service) Register(ctx context.Context, req *model.CreateUserRequest) (*model.UserWithProfile, error) {
	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("username, email, and password are required")
	}

	// Check if user already exists
	existingUser, err := s.repo.User.GetByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to check existing user, caused by %v", err)
		return nil, errors.New("failed to validate user")
	}
	if existingUser != nil {
		return nil, errors.New("username or email already exists")
	}

	// Check by email
	existingUser, err = s.repo.User.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to check existing email, caused by %v", err)
		return nil, errors.New("failed to validate email")
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to hash password, caused by %v", err)
		return nil, errors.New("failed to process password")
	}

	// Create user
	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Phone:        &req.Phone,
		Status:       model.UserStatusActive,
	}

	err = s.repo.User.Create(ctx, user)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create user, caused by %v", err)
		return nil, errors.New("failed to create user account")
	}

	// Create default profile
	profile := &model.Profile{
		ID:       uuid.New(),
		UserID:   user.ID,
		Timezone: "UTC",
		Language: "en",
	}

	err = s.repo.Profile.Create(ctx, profile)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create profile, caused by %v", err)
		// Continue even if profile creation fails
	}

	// Send email verification
	err = s.SendEmailVerification(ctx, user.ID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to send email verification, caused by %v", err)
		// Continue even if email verification fails
	}

	userWithProfile := &model.UserWithProfile{
		User:    *user,
		Profile: profile,
	}

	s.logger.Infofctx(provider.AppLog, ctx, "User registered successfully user_id %s, username %s", user.ID, user.Username)
	return userWithProfile, nil
}

// Login authenticates a user and returns tokens
func (s *service) Login(ctx context.Context, req *model.LoginRequest, clientIP net.IP, userAgent string) (*model.LoginResponse, error) {
	// Get user
	user, err := s.repo.User.GetByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user, caused by %v", err)
		return nil, errors.New("invalid credentials")
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check account status
	if !user.IsActive() {
		return nil, errors.New("account is not active")
	}

	if user.IsLocked() {
		return nil, errors.New("account is temporarily locked")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// Increment failed attempts
		if incrementErr := s.repo.User.IncrementFailedAttempts(ctx, user.ID); incrementErr != nil {
			s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to increment failed attempts, caused by %v", incrementErr)
		}
		return nil, errors.New("invalid credentials")
	}

	// Reset failed attempts on successful login
	if resetErr := s.repo.User.ResetFailedAttempts(ctx, user.ID); resetErr != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to reset failed attempts, caused by %v", resetErr)
	}

	// Update login info
	if updateErr := s.repo.User.UpdateLoginInfo(ctx, user.ID, clientIP); updateErr != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update login info, caused by %v", updateErr)
	}

	// Get user profile
	userWithProfile, err := s.repo.Profile.GetUserWithProfile(ctx, user.ID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user with profile, caused by %v", err)
		// Continue with user data only
		userWithProfile = &model.UserWithProfile{User: *user}
	}

	// Generate tokens
	accessToken, err := s.generateToken(user.ID, model.TokenTypeAccess, AccessTokenExpiry, &userAgent, clientIP)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to generate access token, caused by %v", err)
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.generateToken(user.ID, model.TokenTypeRefresh, RefreshTokenExpiry, &userAgent, clientIP)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to generate refresh token, caused by %v", err)
		return nil, errors.New("failed to generate refresh token")
	}

	response := &model.LoginResponse{
		User:         *userWithProfile,
		AccessToken:  accessToken.TokenHash, // Note: In real implementation, return JWT instead of hash
		RefreshToken: refreshToken.TokenHash,
		ExpiresAt:    accessToken.ExpiresAt,
	}

	s.logger.Infofctx(provider.AppLog, ctx, "User logged in successfully user_id %s, username %s", user.ID, user.Username)
	return response, nil
}

// RefreshToken generates new access token using refresh token
func (s *service) RefreshToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.LoginResponse, error) {
	// Get token with user
	token, user, err := s.repo.Token.GetTokenWithUser(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get refresh token, caused by %v", err)
		return nil, errors.New("invalid refresh token")
	}

	if token == nil || user == nil {
		return nil, errors.New("invalid refresh token")
	}

	if token.TokenType != model.TokenTypeRefresh {
		return nil, errors.New("invalid token type")
	}

	if !token.IsValid() || !user.IsActive() {
		return nil, errors.New("token expired or user inactive")
	}

	// Get user profile
	userWithProfile, err := s.repo.Profile.GetUserWithProfile(ctx, user.ID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user with profile, caused by %v", err)
		userWithProfile = &model.UserWithProfile{User: *user}
	}

	// Generate new access token
	accessToken, err := s.generateToken(user.ID, model.TokenTypeAccess, AccessTokenExpiry, token.UserAgent, *token.LastUsedIP)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to generate new access token, caused by %v", err)
		return nil, errors.New("failed to refresh token")
	}

	response := &model.LoginResponse{
		User:         *userWithProfile,
		AccessToken:  accessToken.TokenHash,
		RefreshToken: req.RefreshToken, // Keep the same refresh token
		ExpiresAt:    accessToken.ExpiresAt,
	}

	s.logger.Infofctx(provider.AppLog, ctx, "Token refreshed successfully user_id %s", user.ID)
	return response, nil
}

// Logout revokes the current token
func (s *service) Logout(ctx context.Context, tokenHash string) error {
	token, err := s.repo.Token.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get token for logout, caused by %v", err)
		return errors.New("invalid token")
	}

	if token == nil {
		return errors.New("token not found")
	}

	err = s.repo.Token.RevokeToken(ctx, token.ID, &token.UserID, "user logout")
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke token, caused by %v", err)
		return errors.New("failed to logout")
	}

	s.logger.Infofctx(provider.AppLog, ctx, "User logged out successfully user_id %s", token.UserID)
	return nil
}

// LogoutAll revokes all user tokens
func (s *service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	err := s.repo.Token.RevokeAllUserTokens(ctx, userID, &userID, "logout from all devices")
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke all tokens, caused by %v", err)
		return errors.New("failed to logout from all devices")
	}

	s.logger.Infofctx(provider.AppLog, ctx, "User logged out from all devices user_id %s", userID)
	return nil
}

// ChangePassword changes user password
func (s *service) ChangePassword(ctx context.Context, userID uuid.UUID, req *model.ChangePasswordRequest) error {
	// Get user
	user, err := s.repo.User.GetByID(ctx, userID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user for password change, caused by %v", err)
		return errors.New("user not found")
	}

	if user == nil {
		return errors.New("user not found")
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to hash new password, caused by %v", err)
		return errors.New("failed to process new password")
	}

	// Update password
	err = s.repo.User.UpdatePassword(ctx, userID, string(newPasswordHash))
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update password, caused by %v", err)
		return errors.New("failed to update password")
	}

	// Revoke all existing tokens to force re-login
	err = s.repo.Token.RevokeAllUserTokens(ctx, userID, &userID, "password changed")
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke tokens after password change, caused by %v", err)
		// Continue anyway
	}

	s.logger.Infofctx(provider.AppLog, ctx, "Password changed successfully user_id %s", userID)
	return nil
}

// ForgotPassword initiates password reset process
func (s *service) ForgotPassword(ctx context.Context, req *model.ResetPasswordRequest) error {
	// Get user by email
	user, err := s.repo.User.GetByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user for password reset, caused by %v", err)
		return errors.New("email not found")
	}

	if user == nil {
		// Don't reveal if email exists or not
		s.logger.Infofctx(provider.AppLog, ctx, "Password reset requested for non-existent email %s", req.Email)
		return nil
	}

	if !user.IsActive() {
		return errors.New("account is not active")
	}

	// Generate reset token
	resetToken := s.generateRandomToken()
	resetTokenHash := s.hashToken(resetToken)

	// Create password reset record
	passwordReset := &model.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: resetTokenHash,
		ExpiresAt: time.Now().Add(ResetTokenExpiry),
	}

	err = s.repo.PasswordReset.Create(ctx, passwordReset)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create password reset, caused by %v", err)
		return errors.New("failed to process password reset request")
	}

	// TODO: Send reset email with resetToken
	// For now, just log it (in production, send email)
	s.logger.Infofctx(provider.AppLog, ctx, "Password reset token generated user_id %s, email %s, token %s", user.ID, user.Email, resetToken)

	return nil
}

// ResetPassword completes password reset process
func (s *service) ResetPassword(ctx context.Context, req *model.ConfirmResetPasswordRequest, clientIP net.IP, userAgent string) error {
	resetTokenHash := s.hashToken(req.Token)

	// Get password reset record
	passwordReset, err := s.repo.PasswordReset.GetByTokenHash(ctx, resetTokenHash)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get password reset, caused by %v", err)
		return errors.New("invalid reset token")
	}

	if passwordReset == nil {
		return errors.New("invalid reset token")
	}

	if !passwordReset.IsValid() {
		return errors.New("reset token is expired or already used")
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to hash new password, caused by %v", err)
		return errors.New("failed to process new password")
	}

	// Update password
	err = s.repo.User.UpdatePassword(ctx, passwordReset.UserID, string(newPasswordHash))
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update password, caused by %v", err)
		return errors.New("failed to reset password")
	}

	// Mark reset as used
	err = s.repo.PasswordReset.MarkAsUsed(ctx, passwordReset.ID, clientIP, userAgent)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to mark reset as used, caused by %v", err)
		// Continue anyway
	}

	// Revoke all existing tokens
	err = s.repo.Token.RevokeAllUserTokens(ctx, passwordReset.UserID, &passwordReset.UserID, "password reset")
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke tokens after password reset, caused by %v", err)
		// Continue anyway
	}

	s.logger.Infofctx(provider.AppLog, ctx, "Password reset successfully user_id %s", passwordReset.UserID)
	return nil
}

// SendEmailVerification sends email verification token
func (s *service) SendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.repo.User.GetByID(ctx, userID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user for email verification, caused by %v", err)
		return errors.New("user not found")
	}

	if user == nil {
		return errors.New("user not found")
	}

	if user.IsEmailVerified() {
		return errors.New("email already verified")
	}

	// Generate verification token
	verificationToken := s.generateRandomToken()
	verificationTokenHash := s.hashToken(verificationToken)

	// Create verification token record
	tokenAccess := &model.TokenAccess{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: verificationTokenHash,
		TokenType: model.TokenTypeVerification,
		ExpiresAt: time.Now().Add(VerificationTokenExpiry),
	}

	err = s.repo.Token.Create(ctx, tokenAccess)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to create verification token, caused by %v", err)
		return errors.New("failed to send verification email")
	}

	// TODO: Send verification email with verificationToken
	// For now, just log it (in production, send email)
	s.logger.Infofctx(provider.AppLog, ctx, "Email verification token generated user_id %s, email %s, token %s", userID, user.Email, verificationToken)

	return nil
}

// VerifyEmail verifies user email with token
func (s *service) VerifyEmail(ctx context.Context, token string) error {
	tokenHash := s.hashToken(token)

	// Get verification token
	tokenAccess, err := s.repo.Token.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get verification token, caused by %v", err)
		return errors.New("invalid verification token")
	}

	if tokenAccess == nil {
		return errors.New("invalid verification token")
	}

	if tokenAccess.TokenType != model.TokenTypeVerification {
		return errors.New("invalid token type")
	}

	if !tokenAccess.IsValid() {
		return errors.New("verification token expired")
	}

	// Get user
	user, err := s.repo.User.GetByID(ctx, tokenAccess.UserID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user for email verification, caused by %v", err)
		return errors.New("user not found")
	}

	if user.IsEmailVerified() {
		return errors.New("email already verified")
	}

	// Mark email as verified
	now := time.Now()
	user.EmailVerifiedAt = &now
	err = s.repo.User.Update(ctx, user)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update email verification, caused by %v", err)
		return errors.New("failed to verify email")
	}

	// Revoke verification token
	err = s.repo.Token.RevokeToken(ctx, tokenAccess.ID, &tokenAccess.UserID, "email verified")
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke verification token, caused by %v", err)
		// Continue anyway
	}

	s.logger.Infofctx(provider.AppLog, ctx, "Email verified successfully user_id %s, email %s", user.ID, user.Email)
	return nil
}

// LoginWithGoogle handles Google OAuth login
func (s *service) LoginWithGoogle(ctx context.Context, googleToken string, clientIP net.IP, userAgent string) (*model.LoginResponse, error) {
	// TODO: Implement Google OAuth token validation
	// This is a placeholder implementation

	// In real implementation:
	// 1. Validate Google token with Google API
	// 2. Extract user info (email, name, etc.)
	// 3. Check if user exists or create new user
	// 4. Generate tokens and return response

	return nil, errors.New("Google OAuth not implemented yet")
}

// ValidateToken validates access token and returns user
func (s *service) ValidateToken(ctx context.Context, tokenHash string, clientIP net.IP) (*model.User, error) {
	token, user, err := s.repo.Token.GetTokenWithUser(ctx, tokenHash)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to validate token, caused by %v", err)
		return nil, errors.New("invalid token")
	}

	if token == nil || user == nil {
		return nil, errors.New("invalid token")
	}

	if token.TokenType != model.TokenTypeAccess {
		return nil, errors.New("invalid token type")
	}

	if !token.IsValid() || !user.IsActive() {
		return nil, errors.New("token expired or user inactive")
	}

	// Update last used
	if updateErr := s.repo.Token.UpdateLastUsed(ctx, token.ID, clientIP); updateErr != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to update token last used, caused by %v", updateErr)
		// Continue anyway
	}

	return user, nil
}

// RevokeToken revokes a specific token
func (s *service) RevokeToken(ctx context.Context, tokenHash string) error {
	token, err := s.repo.Token.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get token for revocation, caused by %v", err)
		return errors.New("invalid token")
	}

	if token == nil {
		return errors.New("token not found")
	}

	err = s.repo.Token.RevokeToken(ctx, token.ID, &token.UserID, "token revoked")
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to revoke token, caused by %v", err)
		return errors.New("failed to revoke token")
	}

	s.logger.Infofctx(provider.AppLog, ctx, "Token revoked successfully user_id %s, token_id %s", token.UserID, token.ID)
	return nil
}

// GetUserProfile gets user profile
func (s *service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.UserWithProfile, error) {
	userWithProfile, err := s.repo.Profile.GetUserWithProfile(ctx, userID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user profile, caused by %v", err)
		return nil, errors.New("failed to get user profile")
	}

	if userWithProfile == nil {
		return nil, errors.New("user not found")
	}

	return userWithProfile, nil
}

// UpdateProfile updates user profile
func (s *service) UpdateProfile(ctx context.Context, userID uuid.UUID, req *model.UpdateProfileRequest) (*model.Profile, error) {
	// Get existing profile
	profile, err := s.repo.Profile.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get profile for update, caused by %v", err)
		return nil, errors.New("failed to get profile")
	}

	// Create profile if doesn't exist
	if profile == nil {
		profile = &model.Profile{
			ID:       uuid.New(),
			UserID:   userID,
			Timezone: "UTC",
			Language: "en",
		}
	}

	// Update fields
	if req.FirstName != nil {
		profile.FirstName = req.FirstName
	}
	if req.LastName != nil {
		profile.LastName = req.LastName
	}
	if req.DisplayName != nil {
		profile.DisplayName = req.DisplayName
	}
	if req.Bio != nil {
		profile.Bio = req.Bio
	}
	if req.Gender != nil {
		profile.Gender = req.Gender
	}
	if req.Country != nil {
		profile.Country = req.Country
	}
	if req.State != nil {
		profile.State = req.State
	}
	if req.City != nil {
		profile.City = req.City
	}
	if req.Address != nil {
		profile.Address = req.Address
	}
	if req.PostalCode != nil {
		profile.PostalCode = req.PostalCode
	}
	if req.Timezone != nil {
		profile.Timezone = *req.Timezone
	}
	if req.Language != nil {
		profile.Language = *req.Language
	}
	if req.WebsiteURL != nil {
		profile.WebsiteURL = req.WebsiteURL
	}

	// Parse date of birth if provided
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		if dob, parseErr := time.Parse("2006-01-02", *req.DateOfBirth); parseErr == nil {
			profile.DateOfBirth = &dob
		}
	}

	// Update profile
	if profile.CreatedAt.IsZero() {
		// Create new profile
		err = s.repo.Profile.Create(ctx, profile)
	} else {
		// Update existing profile
		err = s.repo.Profile.Update(ctx, profile)
	}

	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to save profile, caused by %v", err)
		return nil, errors.New("failed to update profile")
	}

	s.logger.Infofctx(provider.AppLog, ctx, "Profile updated successfully user_id %s", userID)
	return profile, nil
}

// GetUserPermissions gets user permissions
func (s *service) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*model.Permission, error) {
	permissions, err := s.repo.Permission.GetUserPermissions(ctx, userID)
	if err != nil {
		s.logger.Errorfctx(provider.AppLog, ctx, false, "Failed to get user permissions, caused by %v", err)
		return nil, errors.New("failed to get permissions")
	}

	return permissions, nil
}

// Helper functions

func (s *service) generateToken(userID uuid.UUID, tokenType string, expiry time.Duration, userAgent *string, clientIP net.IP) (*model.TokenAccess, error) {
	token := s.generateRandomToken()
	tokenHash := s.hashToken(token)

	tokenAccess := &model.TokenAccess{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		TokenType: tokenType,
		ExpiresAt: time.Now().Add(expiry),
		UserAgent: userAgent,
	}

	if clientIP != nil {
		tokenAccess.LastUsedIP = &clientIP
	}

	err := s.repo.Token.Create(context.Background(), tokenAccess)
	if err != nil {
		return nil, err
	}

	// Note: In real implementation, you might want to return JWT instead of storing in DB
	return tokenAccess, nil
}

func (s *service) generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *service) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
