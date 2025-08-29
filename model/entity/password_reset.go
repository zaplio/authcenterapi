package entity

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// PasswordReset represents the authentication.password_resets table
type PasswordReset struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	Email        string     `db:"email" json:"email"`
	TokenHash    string     `db:"token_hash" json:"-"` // Never expose token hash
	ExpiresAt    time.Time  `db:"expires_at" json:"expires_at"`
	Used         bool       `db:"used" json:"used"`
	UsedAt       *time.Time `db:"used_at" json:"used_at,omitempty"`
	UsedIP       *net.IP    `db:"used_ip" json:"used_ip,omitempty"`
	UserAgent    *string    `db:"user_agent" json:"user_agent,omitempty"`
	Attempts     int        `db:"attempts" json:"attempts"`
	MaxAttempts  int        `db:"max_attempts" json:"max_attempts"`
	BlockedUntil *time.Time `db:"blocked_until" json:"blocked_until,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// ResetPasswordRequest represents the request payload for password reset
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ConfirmResetPasswordRequest represents the request payload for confirming password reset
type ConfirmResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// PasswordReset helper methods

// IsExpired checks if the password reset token is expired
func (pr *PasswordReset) IsExpired() bool {
	return pr.ExpiresAt.Before(time.Now())
}

// IsBlocked checks if the password reset is currently blocked due to too many attempts
func (pr *PasswordReset) IsBlocked() bool {
	return pr.BlockedUntil != nil && pr.BlockedUntil.After(time.Now())
}

// IsValid checks if the password reset token is valid
func (pr *PasswordReset) IsValid() bool {
	return !pr.Used && !pr.IsExpired() && !pr.IsBlocked()
}
