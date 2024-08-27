package utils

import (
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
