package tools

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtClaim struct {
	ID int `json:"id"`
	jwt.StandardClaims
}

// Return: token, expired at
func TokenCreate(id int) (string, time.Time, error) {
	jwtKey := []byte(os.Getenv("JWT_KEY"))
	expiredAt := time.Now().UTC().Add(24 * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JwtClaim{
		ID: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredAt.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	})

	signedStr, err := token.SignedString(jwtKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedStr, expiredAt, nil
}

func TokenValidate(t string) (*jwt.Token, error) {
	var jwtKey = []byte(os.Getenv("JWT_KEY"))
	token, _ := jwt.ParseWithClaims(t, &JwtClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error decoding token")
		}

		return jwtKey, nil
	})

	return token, nil
}
