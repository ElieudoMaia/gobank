package main

import (
	"os"

	jwt "github.com/golang-jwt/jwt/v5"
)

var secret = os.Getenv("JWT_SECRET")

func GenerateJWTToken(account *Account) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":            account.ID,
		"accountNumber": account.Number,
		"accountUser":   account.FirstName,
	})

	hmacSecret := []byte(secret)
	tokenString, err := token.SignedString(hmacSecret)

	return tokenString, err
}
