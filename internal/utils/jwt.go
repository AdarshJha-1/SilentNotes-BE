package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(userId string) interface{} {
	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claim.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil
	}
	return token
}

func VerifyJWT(token string) (jwt.MapClaims, error) {

	verifiedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if !verifiedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claim, _ := verifiedToken.Claims.(jwt.MapClaims)

	return claim, nil
}
