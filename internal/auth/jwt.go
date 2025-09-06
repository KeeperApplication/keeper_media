package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func parsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded public key: %w", err)
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("key type is not *rsa.PublicKey")
	}
}

func ValidateToken(tokenString string, publicKeyPEM string) (string, error) {
	rsaPublicKey, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to load public key for validation: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		return "", fmt.Errorf("invalid token or claims")
	}

	username, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("subject (sub) claim is missing or not a string")
	}
	return username, nil
}
