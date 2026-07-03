package handlers

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateJWT creates a short-lived access token, valid for 15 minutes.
func generateJWT(userID string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// generateSignupToken creates a short-lived token proving an
// email was verified, used only to finish signup.
func generateSignupToken(email string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"email":   email,
		"purpose": "signup",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// verifySignupToken validates a signup token and returns
// the email it was issued for.
func verifySignupToken(tokenString string, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	if claims["purpose"] != "signup" {
		return "", errors.New("wrong token type")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", errors.New("missing email in token")
	}

	return email, nil
}
