package auth

import (
	"github.com/dgrijalva/jwt-go"
	"fmt"
	"time"
)

func GenerateToken() (string, error){

	authKey := "DrogoAuthorizationTokenKey"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "username",
		"date": time.Date(2018, 9, 1, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenString, err := token.SignedString([]byte(authKey))
	if err != nil {
		fmt.Println("error in token", err)
		return "", err
	}

	return tokenString, nil
}