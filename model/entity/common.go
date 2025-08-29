package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB represents a PostgreSQL JSONB field
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

// Composite types for repository relationships

// TokenWithUser represents a token with associated user information
type TokenWithUser struct {
	Token TokenAccess `json:"token"`
	User  User        `json:"user"`
}

// PasswordResetWithUser represents a password reset with associated user information
type PasswordResetWithUser struct {
	PasswordReset PasswordReset `json:"password_reset"`
	User          User          `json:"user"`
}

// PermissionWithUser represents a permission with associated user information
type PermissionWithUser struct {
	Permission Permission `json:"permission"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
}
