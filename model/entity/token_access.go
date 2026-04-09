package entity

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// TokenAccess represents the authentication.token_access table
type TokenAccess struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	TokenHash    string     `db:"token_hash" json:"-"` // Never expose token hash
	TokenType    string     `db:"token_type" json:"token_type"`
	TokenName    *string    `db:"token_name" json:"token_name,omitempty"`
	Scope        *string    `db:"scope" json:"scope,omitempty"`
	ExpiresAt    time.Time  `db:"expires_at" json:"expires_at"`
	LastUsedAt   *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	LastUsedIP   *net.IP    `db:"last_used_ip" json:"last_used_ip,omitempty"`
	UserAgent    *string    `db:"user_agent" json:"user_agent,omitempty"`
	DeviceInfo   JSONB      `db:"device_info" json:"device_info,omitempty"`
	Revoked      bool       `db:"revoked" json:"revoked"`
	RevokedAt    *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	RevokedBy    *uuid.UUID `db:"revoked_by" json:"revoked_by,omitempty"`
	RevokeReason *string    `db:"revoke_reason" json:"revoke_reason,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// RefreshTokenRequest represents the request payload for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenAccess helper methods

// IsExpired checks if the token is expired
func (t *TokenAccess) IsExpired() bool {
	return t.ExpiresAt.Before(time.Now())
}

// IsValid checks if the token is valid (not expired and not revoked)
func (t *TokenAccess) IsValid() bool {
	return !t.IsExpired() && !t.Revoked
}
