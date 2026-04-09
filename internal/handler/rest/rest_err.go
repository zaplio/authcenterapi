package rest

// Error codes
const (
	ErrUnauthorized        = "401"
	ErrBadRequest          = "400"
	ErrNotFound            = "404"
	ErrInternalServerError = "500"
)

// Error messages
const (
	MsgUnauthorized            = "Unauthorized"
	MsgInvalidJSON             = "Invalid JSON format"
	MsgValidationError         = "Validation error"
	MsgUserNotFound            = "User not found"
	MsgInvalidCredentials      = "Invalid credentials"
	MsgTokenRequired           = "Authorization token required"
	MsgInvalidToken            = "Invalid or expired token"
	MsgNoTokenProvided         = "No token provided"
	MsgPasswordResetFailed     = "Password reset failed"
	MsgEmailVerificationFailed = "Email verification failed"
)

// Success messages
const (
	MsgUserRegistered        = "User registered successfully"
	MsgLoginSuccess          = "Login successful"
	MsgTokenRefreshed        = "Token refreshed successfully"
	MsgPasswordResetSent     = "Password reset instructions sent to your email"
	MsgPasswordResetSuccess  = "Password reset successfully"
	MsgEmailVerified         = "Email verified successfully"
	MsgVerificationEmailSent = "Verification email sent"
	MsgGoogleLoginSuccess    = "Google login successful"
	MsgProfileRetrieved      = "Profile retrieved successfully"
	MsgProfileUpdated        = "Profile updated successfully"
	MsgPasswordChanged       = "Password changed successfully"
	MsgLoggedOut             = "Logged out successfully"
	MsgLoggedOutAll          = "Logged out from all devices"
	MsgPermissionsRetrieved  = "Permissions retrieved successfully"
	MsgTokenRevoked          = "Token revoked successfully"
	MsgTokenValid            = "Token is valid"
)
