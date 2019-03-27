package utils

import (
	"crypto/rand"
	"errors"
	"log"

	jwt "github.com/dgrijalva/jwt-go"
)

// Token expiration durations
const (
	StandardTokenExpiration = 15             // StandardTokenExpiration minutes
	ExtendedTokenExpiration = (24 * 30) * 60 //30 days in minutes
	SecretLength            = 128 / 8        // 128 bits
)

var jwtSecret []byte = nil

// Generate and use a new secret. The returned secret should be stored
// by the caller and passed to the LoadJWTSecret on each startup.
func GenerateJWTSecret() ([]byte, error) {
	secret := make([]byte, SecretLength)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}
	jwtSecret = secret
	return secret, nil
}

// Use the given secret for generating and validating tokens
func LoadJWTSecret(secret []byte) {
	jwtSecret = secret
}

// GenerateJWTToken returns a JWT token based on the uuid, expiration and the secret sign
func GenerateJWTToken(uuid string, expiration int64) (string, error) {
	if jwtSecret == nil || len(jwtSecret) != SecretLength {
		return "", errors.New("No JWT secret set")
	}

	claims := jwt.StandardClaims{
		ExpiresAt: expiration,
		Id:        uuid}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return s, nil
}

// ValidateJWTToken Returns if the given token is a valid token
func ValidateJWTToken(tokenString string) (*jwt.StandardClaims, error) {
	if jwtSecret == nil || len(jwtSecret) != SecretLength {
		return nil, errors.New("No JWT secret set")
	}
	log.Println("using secret", jwtSecret)

	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if jwtSecret == nil || len(jwtSecret) == 0 {
			return nil, errors.New("No JWT secret set")
		}
		return jwtSecret, nil
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
