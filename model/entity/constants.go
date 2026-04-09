package entity

// TokenType constants
const (
	TokenTypeAccess       = "access"
	TokenTypeRefresh      = "refresh"
	TokenTypeAPIKey       = "api_key"
	TokenTypeVerification = "verification"
)

// User status constants
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusDeleted   = "deleted"
)

// Gender constants
const (
	GenderMale           = "male"
	GenderFemale         = "female"
	GenderOther          = "other"
	GenderPreferNotToSay = "prefer_not_to_say"
)

// Common permission constants
const (
	PermissionAll       = "*"
	PermissionRead      = "read"
	PermissionWrite     = "write"
	PermissionDelete    = "delete"
	PermissionAdmin     = "admin"
	PermissionModerator = "moderator"
)

// Common role constants
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleModerator  = "moderator"
	RoleUser       = "user"
	RoleGuest      = "guest"
)
