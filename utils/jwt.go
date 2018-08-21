package utils

import (
	"os"

	jwt "github.com/dgrijalva/jwt-go"
)

// Token expiration durations
const (
	StandardTokenExpiration = 15             // StandardTokenExpiration minutes
	ExtendedTokenExpiration = (24 * 30) * 60 //30 days in minutes
)

// GenerateJWTToken returns a JWT token based on the uuid, expiration and the secret sign
func GenerateJWTToken(uuid string, expiration int64) (string, error) {
	claims := jwt.StandardClaims{
		ExpiresAt: expiration,
		Id:        uuid}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return s, nil
}

// ValidateJWTToken Returns if the given token is a valid token
func ValidateJWTToken(tokenString string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
