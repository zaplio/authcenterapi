package rest

import (
	"authcenterapi/model"
	"authcenterapi/model/entity"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Authentication endpoints

// Register creates a new user account
func (rs *Rest) Register(c *gin.Context) {
	var req entity.CreateUserRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	user, err := rs.service.Register(c.Request.Context(), &req)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusCreated, MsgUserRegistered, user)
}

// Login authenticates user and returns tokens
func (rs *Rest) Login(c *gin.Context) {
	var req entity.LoginRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	clientIP := rs.getClientIP(c)
	userAgent := c.GetHeader("User-Agent")

	response, err := rs.service.Login(c.Request.Context(), &req, clientIP, userAgent)
	if err != nil {
		rs.errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgLoginSuccess, response)
}

// RefreshToken generates new access token
func (rs *Rest) RefreshToken(c *gin.Context) {
	var req entity.RefreshTokenRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	response, err := rs.service.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		rs.errorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgTokenRefreshed, response)
}

// ForgotPassword initiates password reset
func (rs *Rest) ForgotPassword(c *gin.Context) {
	var req entity.ResetPasswordRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	err := rs.service.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgPasswordResetSent, nil)
}

// ResetPassword completes password reset
func (rs *Rest) ResetPassword(c *gin.Context) {
	var req entity.ConfirmResetPasswordRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	clientIP := rs.getClientIP(c)
	userAgent := c.GetHeader("User-Agent")

	err := rs.service.ResetPassword(c.Request.Context(), &req, clientIP, userAgent)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgPasswordResetSuccess, nil)
}

// VerifyEmail verifies user email with token
func (rs *Rest) VerifyEmail(c *gin.Context) {
	type VerifyEmailRequest struct {
		Token string `json:"token" validate:"required"`
	}

	var req VerifyEmailRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	err := rs.service.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgEmailVerified, nil)
}

// ResendEmailVerification resends email verification
func (rs *Rest) ResendEmailVerification(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		rs.errorResponse(c, http.StatusUnauthorized, MsgUnauthorized)
		return
	}

	err := rs.service.SendEmailVerification(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgVerificationEmailSent, nil)
}

// LoginWithGoogle handles Google OAuth login
func (rs *Rest) LoginWithGoogle(c *gin.Context) {
	type GoogleLoginRequest struct {
		Token string `json:"token" validate:"required"`
	}

	var req GoogleLoginRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	clientIP := rs.getClientIP(c)
	userAgent := c.GetHeader("User-Agent")

	response, err := rs.service.LoginWithGoogle(c.Request.Context(), req.Token, clientIP, userAgent)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgGoogleLoginSuccess, response)
}

// Protected endpoints (require authentication)

// GetProfile gets user profile
func (rs *Rest) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		rs.errorResponse(c, http.StatusUnauthorized, MsgUnauthorized)
		return
	}

	profile, err := rs.service.GetUserProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		rs.errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgProfileRetrieved, profile)
}

// UpdateProfile updates user profile
func (rs *Rest) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		rs.errorResponse(c, http.StatusUnauthorized, MsgUnauthorized)
		return
	}

	var req entity.UpdateProfileRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	profile, err := rs.service.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgProfileUpdated, profile)
}

// ChangePassword changes user password
func (rs *Rest) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		rs.errorResponse(c, http.StatusUnauthorized, MsgUnauthorized)
		return
	}

	var req entity.ChangePasswordRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	err := rs.service.ChangePassword(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgPasswordChanged, nil)
}

// Logout logs out current session
func (rs *Rest) Logout(c *gin.Context) {
	token := rs.extractToken(c)
	if token == "" {
		rs.errorResponse(c, http.StatusUnauthorized, MsgNoTokenProvided)
		return
	}

	err := rs.service.Logout(c.Request.Context(), token)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgLoggedOut, nil)
}

// LogoutAll logs out from all devices
func (rs *Rest) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		rs.errorResponse(c, http.StatusUnauthorized, MsgUnauthorized)
		return
	}

	err := rs.service.LogoutAll(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgLoggedOutAll, nil)
}

// GetUserPermissions gets user permissions
func (rs *Rest) GetUserPermissions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		rs.errorResponse(c, http.StatusUnauthorized, MsgUnauthorized)
		return
	}

	permissions, err := rs.service.GetUserPermissions(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgPermissionsRetrieved, permissions)
}

// Token management endpoints

// RevokeToken revokes a specific token
func (rs *Rest) RevokeToken(c *gin.Context) {
	type RevokeTokenRequest struct {
		Token string `json:"token" validate:"required"`
	}

	var req RevokeTokenRequest
	if err := rs.bindAndValidate(c, &req); err != nil {
		return
	}

	err := rs.service.RevokeToken(c.Request.Context(), req.Token)
	if err != nil {
		rs.errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	rs.successResponse(c, http.StatusOK, MsgTokenRevoked, nil)
}

// ValidateToken validates current token
func (rs *Rest) ValidateToken(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		rs.errorResponse(c, http.StatusUnauthorized, MsgUnauthorized)
		return
	}

	rs.successResponse(c, http.StatusOK, MsgTokenValid, user)
}

// Middleware functions

// authMiddleware validates JWT token
func (rs *Rest) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := rs.extractToken(c)
		if token == "" {
			rs.errorResponse(c, http.StatusUnauthorized, MsgTokenRequired)
			c.Abort()
			return
		}

		clientIP := rs.getClientIP(c)
		user, err := rs.service.ValidateToken(c.Request.Context(), token, clientIP)
		if err != nil {
			rs.errorResponse(c, http.StatusUnauthorized, MsgInvalidToken)
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", user.ID)
		c.Set("user", user)
		c.Set("token", token)

		c.Next()
	}
}

// corsMiddleware adds CORS headers
func (rs *Rest) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// Helper functions

func (rs *Rest) bindAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		rs.errorResponse(c, http.StatusBadRequest, MsgInvalidJSON)
		return err
	}

	if err := rs.validator.Struct(obj); err != nil {
		rs.errorResponse(c, http.StatusBadRequest, fmt.Sprintf("Validation error: %v", err))
		return err
	}

	return nil
}

func (rs *Rest) extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if bearerToken == "" {
		return ""
	}

	// Extract token from "Bearer <token>" format
	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func (rs *Rest) getClientIP(c *gin.Context) net.IP {
	// Try to get real IP from headers
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		ip := net.ParseIP(xRealIP)
		if ip != nil {
			return ip
		}
	}

	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// Take the first IP in the list
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := net.ParseIP(strings.TrimSpace(ips[0]))
			if ip != nil {
				return ip
			}
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return net.ParseIP("127.0.0.1")
	}

	return net.ParseIP(ip)
}

func (rs *Rest) errorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, model.ApiResponse{
		Status: false,
		Error: &model.Error{
			Code:    fmt.Sprintf("%d", statusCode),
			Message: message,
		},
	})
}

func (rs *Rest) successResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, model.ApiResponse{
		Status:  true,
		Message: message,
		Data:    data,
	})
}
