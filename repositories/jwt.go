package repositories

import (
	"jc-financas/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	TeamID uint `json:"teamId"`
	UserID uint `json:"userId"`
	jwt.RegisteredClaims
}

func GenerateJWT(user models.User) (string, error) {
	claims := MyCustomClaims{
		UserID: user.ID,
		TeamID: user.TeamID,
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
