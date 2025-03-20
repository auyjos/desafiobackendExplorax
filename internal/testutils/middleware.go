package testutils

import (
	"github.com/gin-gonic/gin"
)

// MockJWTMiddleware returns a simplified middleware for testing
func MockJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	}
}

// SetupTestRouter returns a gin engine with test middleware configured
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(MockJWTMiddleware())
	return r
}
