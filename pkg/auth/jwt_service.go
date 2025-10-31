package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secretKey     []byte
	tokenLifespan time.Duration
}

type CustomClaims struct {
	OwnerID uuid.UUID `json:"owner_id"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKey string, tokenLifespan time.Duration) *JWTService {
	return &JWTService{
		secretKey:     []byte(secretKey),
		tokenLifespan: tokenLifespan,
	}
}

func (s *JWTService) GenerateToken(ownerID uuid.UUID) (string, error) {
	claims := CustomClaims{
		ownerID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenLifespan)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   ownerID.String(),
			Issuer:    "personal-os-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("cannot sign token: %w", err)
	}

	return signedString, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signature algorithm: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("error when parsing token claims")
}
