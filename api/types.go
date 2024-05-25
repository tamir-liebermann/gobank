package api

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
	UserName string  `json:"user_name"`
	Password string `json:"password"`
}

type CreateAccountRequest = LoginRequest

type TransferRequest struct {
	ToAccountID string `json:"to_account_id"`
	Amount    int `json:"amount"`
}
