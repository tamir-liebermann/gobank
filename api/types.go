package api

import (
	"github.com/tamir-liebermann/gobank/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type CreateAccountRequest struct {
	Password    string  `json:"password"`
	UserName    string  `json:"user_name"`
	Balance     float64 `json:"balance"`
	PhoneNumber string  `json:"phone_number"`
}

type TransferRequest struct {
	From   string  `json:"from,omitempty"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}


type AccNameReq struct {
	AccountHolder string `json:"account_holder"`
}

type TransactionsReq struct {
	    AccountID string `json:"_id"`

}
type AllTransactionsRes struct {
    Transactions []db.Transaction `json:"transactions"`
    
}
type ErrorResponse struct {
    Message string `json:"message"`
    // Add other fields as needed
}

type BankAccRes struct {
	BankAcc db.BankAccount `json:"bank_account"`
}

type BankAccsRes struct {
	BankAccs []db.BankAccount
}

type BalanceRequest struct {
    AccountID string `json:"_id" binding:"required"`
	AccountName string `json:"account_holder"`
}

type BalanceResponse struct {
    Balance float64 `json:"balance"`
	Transactions []TransactionInfo `json:"transactions"`
}
type TransactionInfo struct {
    FromAccount string  `json:"from_account"`
    Amount      float64 `json:"amount"`
}



type DepositRequest struct {
	AccountID string `json:"_id"`
	Amount 	  float64 `json:"amount"`
}

type DepositResponse struct {
	Message 	string `json:"message"`
}

type GPTRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type GPTResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

type ChatReq struct {
	UserText string `json:"user_text"`
	From 	string 	`json:"from"`
}

type HealthResponse struct {
    Status string `json:"status"`
}