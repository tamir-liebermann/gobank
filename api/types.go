package api

import "go.mongodb.org/mongo-driver/bson/primitive"

// 	"context"
// 	"log"
// 	"math/rand"
// 	"time"

//	"golang.org/x/crypto/bcrypt"
//
// )

type LoginResponse struct {
	UserName string  `json:"user_name"`
	Token  string `json:"token"`
}

type LoginRequest struct {
	Password string `json:"password"`
	UserName string  `json:"user_name"`
}

type IdRequest struct  {
	Id 	primitive.ObjectID      `json:"_id"`
	
}

type IdResponse = LoginResponse

type CreateAccountResponse struct { 
	Message  string 		    `json:"message"`
	Id       string				`json:"_id"`
	Token  	 string 			`json:"token"`
}

type CreateAccountRequest = LoginRequest

type TransferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID string `json:"to_account_id"`
	Amount    float64 `json:"amount"`
}
