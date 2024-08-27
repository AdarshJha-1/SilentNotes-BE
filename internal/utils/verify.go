package utils

import (
	"math/rand/v2"
	"time"
)

func GenerateVerifyCode() int {

	randomNumber := rand.IntN(900000) + 100000
	return randomNumber
}

func VerifyCodeExpiry() time.Time {
	expiryTime := time.Now().Add((2 * time.Hour))
	return expiryTime
}
