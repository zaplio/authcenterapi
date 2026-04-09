package entity

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents the authentication.permissions table
type Permission struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	Role       string     `db:"role" json:"role"`
	Permission string     `db:"permission" json:"permission"`
	Resource   *string    `db:"resource" json:"resource,omitempty"`
	ResourceID *uuid.UUID `db:"resource_id" json:"resource_id,omitempty"`
	GrantedBy  *uuid.UUID `db:"granted_by" json:"granted_by,omitempty"`
	GrantedAt  time.Time  `db:"granted_at" json:"granted_at"`
	ExpiresAt  *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	Revoked    bool       `db:"revoked" json:"revoked"`
	RevokedAt  *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	RevokedBy  *uuid.UUID `db:"revoked_by" json:"revoked_by,omitempty"`
	Conditions JSONB      `db:"conditions" json:"conditions,omitempty"`
	Metadata   JSONB      `db:"metadata" json:"metadata,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}

// UserPermissionView represents the v_user_permissions view
type UserPermissionView struct {
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	Username   string     `db:"username" json:"username"`
	Email      string     `db:"email" json:"email"`
	Role       string     `db:"role" json:"role"`
	Permission string     `db:"permission" json:"permission"`
	Resource   *string    `db:"resource" json:"resource,omitempty"`
	ResourceID *uuid.UUID `db:"resource_id" json:"resource_id,omitempty"`
	ExpiresAt  *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	Revoked    bool       `db:"revoked" json:"revoked"`
	Conditions JSONB      `db:"conditions" json:"conditions,omitempty"`
	GrantedAt  time.Time  `db:"granted_at" json:"granted_at"`
}

// GrantPermissionRequest represents the request payload for granting permissions
type GrantPermissionRequest struct {
	UserID     uuid.UUID  `json:"user_id" validate:"required"`
	Role       string     `json:"role" validate:"required"`
	Permission string     `json:"permission" validate:"required"`
	Resource   *string    `json:"resource,omitempty"`
	ResourceID *uuid.UUID `json:"resource_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// Permission helper methods

// IsExpired checks if the permission is expired
func (p *Permission) IsExpired() bool {
	return p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now())
}

// IsValid checks if the permission is valid (not expired and not revoked)
func (p *Permission) IsValid() bool {
	return !p.Revoked && !p.IsExpired()
}
