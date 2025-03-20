package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// TestSecret is used for testing purposes
const TestSecret = "test-secret-key"

func init() {
	os.Setenv("JWT_SECRET", TestSecret)
}

// GenerateTestToken creates a JWT token for testing
func GenerateTestToken(userID string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

	tokenString, _ := token.SignedString([]byte(TestSecret))
	return tokenString
}

// MockJWTMiddleware returns a simplified middleware for testing
func MockJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Next()
	}
}

// WithAuthToken is a test helper that adds a JWT token to the context
func WithAuthToken(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		token := GenerateTestToken("test-user")
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		JWTAuthMiddleware()(c)

		if c.Errors != nil {
			t.Errorf("Expected no errors, got %v", c.Errors)
		}

		userID, exists := c.Get("user_id")
		if !exists {
			t.Error("Expected user_id to be set")
		}
		if userID != "test-user" {
			t.Errorf("Expected user_id to be 'test-user', got %v", userID)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		JWTAuthMiddleware()(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %v", w.Code)
		}
	})

	t.Run("Missing token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("GET", "/", nil)

		JWTAuthMiddleware()(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %v", w.Code)
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["user_id"] = "test-user"
		claims["exp"] = time.Now().Add(-24 * time.Hour).Unix()

		tokenString, _ := token.SignedString([]byte(TestSecret))

		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+tokenString)

		JWTAuthMiddleware()(c)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %v", w.Code)
		}
	})
}
