package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const secretKey = "supersecret"

func GenerateToken(username string, userId primitive.ObjectID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"userId":   userId.Hex(),
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	return token.SignedString([]byte(secretKey))
}

func VerifyToken(token string) (primitive.ObjectID, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, errors.New("unexpected sigining method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return primitive.NilObjectID, errors.New("could not parse token")
	}

	tokenIsValid := parsedToken.Valid

	if !tokenIsValid {
		return primitive.NilObjectID, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)

	if !ok {
		return primitive.NilObjectID, errors.New("invalid token claims")
	}
	userIdHex, ok := claims["userId"].(string)
	if !ok {
		return primitive.NilObjectID, errors.New("invalid user ID in token claims")
	}

	userId, err := primitive.ObjectIDFromHex(userIdHex)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid user ID format in token claims")
	}
	return userId, nil
}
