package rest

import (
	"authcenterapi/internal/provider"
	"authcenterapi/internal/service"
	"authcenterapi/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Rest struct {
	log       provider.ILogger
	service   service.IService
	validator *validator.Validate
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewRest(log provider.ILogger, service service.IService) *Rest {
	return &Rest{
		log:       log,
		service:   service,
		validator: validator.New(),
	}
}

func (rs *Rest) CreateServer(address string) (*http.Server, error) {
	gin.SetMode(util.Configuration.Server.Mode)

	r := gin.New()

	// Add CORS middleware
	r.Use(rs.corsMiddleware())

	// Add logging middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Register routes
	rs.registerRoutes(r)

	server := &http.Server{
		Addr:    address,
		Handler: r,
	}

	return server, nil
}

func (rs *Rest) registerRoutes(router *gin.Engine) {
	// Health check
	router.GET("/ping", rs.checkConnectivity)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", rs.Register)
			auth.POST("/login", rs.Login)
			auth.POST("/refresh", rs.RefreshToken)
			auth.POST("/forgot-password", rs.ForgotPassword)
			auth.POST("/reset-password", rs.ResetPassword)
			auth.POST("/verify-email", rs.VerifyEmail)

			// OAuth routes
			auth.POST("/google", rs.LoginWithGoogle)
		}

		// Protected routes (require authentication)
		protected := v1.Group("/")
		protected.Use(rs.authMiddleware())
		{
			// User management
			user := protected.Group("/user")
			{
				user.GET("/profile", rs.GetProfile)
				user.PUT("/profile", rs.UpdateProfile)
				user.POST("/change-password", rs.ChangePassword)
				user.POST("/logout", rs.Logout)
				user.POST("/logout-all", rs.LogoutAll)
				user.GET("/permissions", rs.GetUserPermissions)
				user.POST("/resend-verification", rs.ResendEmailVerification)
			}

			// Token management
			token := protected.Group("/token")
			{
				token.POST("/revoke", rs.RevokeToken)
				token.GET("/validate", rs.ValidateToken)
			}
		}
	}
}

// Health check endpoint
func (rs *Rest) checkConnectivity(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().UTC(),
	})
}
