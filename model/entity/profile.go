package entity

import (
	"time"

	"github.com/google/uuid"
)

// Profile represents the authentication.profiles table
type Profile struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	FirstName   *string    `db:"first_name" json:"first_name,omitempty"`
	LastName    *string    `db:"last_name" json:"last_name,omitempty"`
	DisplayName *string    `db:"display_name" json:"display_name,omitempty"`
	AvatarURL   *string    `db:"avatar_url" json:"avatar_url,omitempty"`
	Bio         *string    `db:"bio" json:"bio,omitempty"`
	DateOfBirth *time.Time `db:"date_of_birth" json:"date_of_birth,omitempty"`
	Gender      *string    `db:"gender" json:"gender,omitempty"`
	Country     *string    `db:"country" json:"country,omitempty"`
	State       *string    `db:"state" json:"state,omitempty"`
	City        *string    `db:"city" json:"city,omitempty"`
	Address     *string    `db:"address" json:"address,omitempty"`
	PostalCode  *string    `db:"postal_code" json:"postal_code,omitempty"`
	Timezone    string     `db:"timezone" json:"timezone"`
	Language    string     `db:"language" json:"language"`
	WebsiteURL  *string    `db:"website_url" json:"website_url,omitempty"`
	SocialLinks JSONB      `db:"social_links" json:"social_links,omitempty"`
	Preferences JSONB      `db:"preferences" json:"preferences,omitempty"`
	Metadata    JSONB      `db:"metadata" json:"metadata,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// UpdateProfileRequest represents the request payload for updating profile information
type UpdateProfileRequest struct {
	FirstName   *string `json:"first_name,omitempty"`
	LastName    *string `json:"last_name,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	DateOfBirth *string `json:"date_of_birth,omitempty"` // Format: YYYY-MM-DD
	Gender      *string `json:"gender,omitempty" validate:"omitempty,oneof=male female other prefer_not_to_say"`
	Country     *string `json:"country,omitempty" validate:"omitempty,len=2"`
	State       *string `json:"state,omitempty"`
	City        *string `json:"city,omitempty"`
	Address     *string `json:"address,omitempty"`
	PostalCode  *string `json:"postal_code,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	Language    *string `json:"language,omitempty" validate:"omitempty,len=2"`
	WebsiteURL  *string `json:"website_url,omitempty" validate:"omitempty,url"`
}

// Profile helper methods

// GetFullName returns the user's full name from profile
func (p *Profile) GetFullName() string {
	if p.FirstName == nil && p.LastName == nil {
		return ""
	}

	var fullName string
	if p.FirstName != nil {
		fullName = *p.FirstName
	}
	if p.LastName != nil {
		if fullName != "" {
			fullName += " "
		}
		fullName += *p.LastName
	}

	return fullName
}
