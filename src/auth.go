package main

import (
	"fmt"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var secret = os.Getenv("JWT_SECRET")

func GenerateJWTToken(account *Account) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID":            account.ID,
		"accountNumber": account.Number,
		"accountUser":   account.FirstName,
		"exp":           time.Now().Add(time.Hour * 1).Unix(),
	})

	hmacSecret := []byte(secret)
	tokenString, err := token.SignedString(hmacSecret)

	return tokenString, err
}

func VerifyToken(tokenString string) (accountID int, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		hmacSecret := []byte(secret)
		return hmacSecret, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		id, ok := claims["ID"].(float64)
		if !ok {
			return 0, fmt.Errorf("ID is not an int")
		}
		return int(id), nil
	} else {
		return 0, fmt.Errorf("error getting token payload")
	}
}
