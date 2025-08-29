package entity

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// User represents the authentication.users table
type User struct {
	ID                  uuid.UUID  `db:"id" json:"id"`
	Username            string     `db:"username" json:"username"`
	Email               string     `db:"email" json:"email"`
	EmailVerifiedAt     *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`
	PasswordHash        string     `db:"password_hash" json:"-"` // Never expose password hash in JSON
	Phone               *string    `db:"phone" json:"phone,omitempty"`
	PhoneVerifiedAt     *time.Time `db:"phone_verified_at" json:"phone_verified_at,omitempty"`
	Status              string     `db:"status" json:"status"`
	LastLoginAt         *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	LastLoginIP         *net.IP    `db:"last_login_ip" json:"last_login_ip,omitempty"`
	FailedLoginAttempts int        `db:"failed_login_attempts" json:"failed_login_attempts"`
	LockedUntil         *time.Time `db:"locked_until" json:"locked_until,omitempty"`
	TwoFactorEnabled    bool       `db:"two_factor_enabled" json:"two_factor_enabled"`
	TwoFactorSecret     *string    `db:"two_factor_secret" json:"-"` // Never expose 2FA secret
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt           *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// UserWithProfile represents a user with their profile information
type UserWithProfile struct {
	User
	Profile *Profile `json:"profile,omitempty"`
}

// CreateUserRequest represents the request payload for creating a new user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,e164"`
}

// UpdateUserRequest represents the request payload for updating user information
type UpdateUserRequest struct {
	Email string `json:"email,omitempty" validate:"omitempty,email"`
	Phone string `json:"phone,omitempty" validate:"omitempty,e164"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the response payload for successful login
type LoginResponse struct {
	User         UserWithProfile `json:"user"`
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

// ChangePasswordRequest represents the request payload for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// User helper methods

// IsActive checks if the user is active and not soft deleted
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// IsLocked checks if the user account is currently locked
func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}

// IsEmailVerified checks if the user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// IsPhoneVerified checks if the user's phone is verified
func (u *User) IsPhoneVerified() bool {
	return u.PhoneVerifiedAt != nil && u.Phone != nil
}
