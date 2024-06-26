package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tamir-liebermann/gobank/env"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GenerateToken(username string, userId primitive.ObjectID) (string, error) {
	spec := env.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"userId":   userId.Hex(),
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString([]byte(spec.JwtSecret))
}

func VerifyToken(token string) (string, error) {
		spec := env.New()

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, errors.New("unexpected sigining method")
		}
		return []byte(spec.JwtSecret), nil
	})

	if err != nil {
		return "", errors.New("could not parse token")
	}

	tokenIsValid := parsedToken.Valid

	if !tokenIsValid {
		return "", errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)

	if !ok {
		return "", errors.New("invalid token claims")
	}
	userIdHex, ok := claims["userId"].(string)
	if !ok {
		return "", errors.New("invalid user ID in token claims")
	}

	return  userIdHex, nil

}
