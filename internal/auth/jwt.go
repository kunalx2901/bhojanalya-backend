package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET not set")
	}
	return []byte(secret), nil
}

func GenerateToken(userID, email, role string) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenString string) (string, string, string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", "", "", err
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})

	if err != nil {
		return "", "", "", errors.New("token parsing failed: " + err.Error())
	}

	if !token.Valid {
		return "", "", "", errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", "", errors.New("invalid claims type")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", "", errors.New("user_id claim not found or invalid type")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", "", "", errors.New("email claim not found or invalid type")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return "", "", "", errors.New("role claim not found or invalid type")
	}

	return userID, email, role, nil
}
