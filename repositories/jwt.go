package repositories

import (
	"jc-financas/models"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type MyCustomClaims struct {
	TeamID    uint   `json:"teamId"`
	UserID    uint   `json:"userId"`
	UserEmail string `json:"userEmail"`
	jwt.RegisteredClaims
}

func GenerateJWT(user models.User) (string, error) {
	claims := MyCustomClaims{
		UserID:    user.ID,
		TeamID:    user.TeamID,
		UserEmail: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "api.orfed.com",
			Subject:   "somebody",
			// ID:        "1",
			Audience: []string{"somebody_else"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// jwtSecret := os.Getenv("JWT_SECRET")
	jwtSecret := "a2tra2s="

	hmacSampleSecret := []byte(jwtSecret)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(hmacSampleSecret)

	return tokenString, err
}

func Parse(stringToken string) (MyCustomClaims, error) {
	claims := MyCustomClaims{}

	jwtSecret := "a2tra2s="

	hmacSampleSecret := []byte(jwtSecret)

	token, err := jwt.ParseWithClaims(stringToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})

	if err != nil {
		return claims, err
	}

	if !token.Valid {
		return claims, err
	}

	return claims, nil
}

func ParseWithContext(c echo.Context) (*MyCustomClaims, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("Token não fornecido")
	}

	claims, err := Parse(strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer")))
	if err != nil {

		return nil, errors.New("Token inválido")
	}

	return &claims, nil
}
