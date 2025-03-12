// /internal/utils/jwt.go
package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

// GenerateJWT genera un token JWT para un user_id dado.
func GenerateJWT(userID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(72 * time.Hour).Unix() // Expira en 72 horas

	secret := os.Getenv("JWT_SECRET")
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
