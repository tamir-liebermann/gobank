package api

import (
	"fmt"
	"log"
	"net/http"

	// "strings"

	"github.com/gin-gonic/gin"
	"github.com/tamir-liebermann/gobank/db"
	"github.com/tamir-liebermann/gobank/utils"

	// openapi "github.com/twilio/twilio-go/rest/accounts/v1"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Call GPT API to generate response based on user input

// @Summary Create a new account
// @Description Create a new account with username and password
// @ID create-account
// @Accept  json
// @Produce  json
// @Param   account  body     CreateAccountRequest  true  "Account Information"
// @Success 201 {object} string "Account created!"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /create [post]
func (api *ApiManager) handleCreateAccount(ctx *gin.Context) {
	var req CreateAccountRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Bad request"})

	}

	account, err := api.accMgr.CreateAccount(req.UserName, req.Password, req.Balance, req.PhoneNumber, req.Role)
	if err != nil {

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Could not create account, try again later"})
		return
	}

	token, err := utils.GenerateToken(req.UserName, account.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generating token "})
		return
	}
	response := CreateAccountResponse{
		Message: "Account created!",
		Id:      account.ID.Hex(),
		Token:   token,
	}
	ctx.Set("UserRole", account.Role)
	ctx.Next()
	ctx.JSON(http.StatusCreated, response)
}

// @Summary Login your account
// @Description login your account via username and password
// @ID login
// @Accept  json
// @Produce  json
// @Param   login  body     LoginRequest  true  "Login Information"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /login [post]
func (api *ApiManager) handleLogin(ctx *gin.Context) {
	var req LoginRequest

	// Parse the JSON request body
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Bad request"})
		return
	}

	// Validate required fields
	if req.UserName == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Username and password are required"})
		return
	}

	// Search for the account by user name (returns a slice of accounts)
	accounts, err := api.accMgr.SearchAccountByNameOrPhone(req.UserName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error"})
		return
	}

	// Check if no accounts were found
	if len(accounts) == 0 {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid credentials"})
		return
	}

	// Iterate over the returned accounts to find a match with the password
	var account *db.BankAccount
	for _, acc := range accounts {
		if utils.CheckPasswordHash(req.Password, acc.Password) {
			account = acc
			break
		}
	}

	// If no account matches the provided password
	if account == nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid credentials"})
		return
	}

	// Generate a token for the authenticated user
	token, err := utils.GenerateToken(account.AccountHolder, account.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generating token"})
		return
	}

	// Prepare and send the response
	response := LoginResponse{
		UserName: req.UserName,
		Token:    token,
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get account by ID
// @Description Get account details by its ID
// @ID get-account-by-id
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} BankAccRes "Account found!"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Account not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/{id} [get]
// @Security BearerAuth
func (api *ApiManager) handleGetById(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid ID format"})
		return
	}

	account, err := api.accMgr.SearchAccountById(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error"})
		return
	}

	if account == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "Account not found"})
		return
	}
	ctx.JSON(http.StatusOK, account)

}

// @Summary Get account by account holder's name
// @Description Retrieve account details by the account holder's name
// @ID get-account-by-name-or-phone
// @Produce json
// @Param account_holder query string false "Account Holder's Name"
// @Success 200 {object} BankAccRes    "Account found!"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Account not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/name/{account_holder} [get]
// @Security BearerAuth
func (api *ApiManager) handleGetByNameOrPhone(ctx *gin.Context) {
	// Extract the account holder's name from the request parameters
	accountHolder := ctx.Query("account_holder")

	if accountHolder == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Query parameter is required"})
		return
	}

	// Call the SearchAccountByNameOrPhone function
	accounts, err := api.handleSearchAccountByNameIntent(accountHolder)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error: " + err.Error()})
		return
	}

	// Prepare and send the response
	ctx.JSON(http.StatusOK, accounts)
}

// @Param id path string true "Account ID"
// @Success 200 {object} string "Success"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Account not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/{id} [delete]
// @Security BearerAuth
func (api *ApiManager) handleDeleteById(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid ID format"})
		return
	}

	err = api.accMgr.DeleteAccountById(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Account Deleted Successfully!"})
}

// @Summary Get all accounts
// @Description Retrieve a list of all accounts
// @ID get-accounts
// @Accept  json
// @Produce  json
// @Success 200 {array} BankAccsRes  "Accounts found"
// @Failure 500 {object} ErrorResponse "Could not retrive accounts"
// @Router /admin/accounts [get]
// @Security BearerAuth
func (api *ApiManager) handleGetAccounts(ctx *gin.Context) {
	// Assume you have set the userId in the context after authentication
	userId, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Message: "User ID not found"})
		return
	}

	// Fetch the account based on userId
	objectID, err := primitive.ObjectIDFromHex(userId.(string))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid user ID format"})
		return
	}

	account, err := api.accMgr.SearchAccountById(objectID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Could not find account"})
		return
	}

	// Check the role of the account
	if account.Role != "admin" {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Message: "You are not authorized to perform this action"})
		return
	}

	// Fetch all accounts
	accounts, err := api.accMgr.GetAccounts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Could not find accounts"})
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

// @Summary Transfer funds from one account to another
// @Description Transfer funds from one bank account to another
// @ID transfer-funds
// @Accept json
// @Produce json
// @Param request body TransferRequest true "Transfer Request"
// @Success 200 {object} BankAccRes
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Invalid account ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/transfer [post]
// @Security BearerAuth
func (api *ApiManager) handleTransfer(ctx *gin.Context) {
	log.Println("transferHandler called")
	var req TransferRequest

	// Use ShouldBindJSON to parse the request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		return
	}
	log.Printf("Received transfer request: From=%s To=%s Amount=%f", req.From, req.To, req.Amount)

	fromAccountID, err := primitive.ObjectIDFromHex(req.From)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid from account ID"})
		return
	}

	toAccountID, err := primitive.ObjectIDFromHex(req.To)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid to account ID"})
		return
	}

	err = api.accMgr.TransferAmountById(fromAccountID, toAccountID, req.Amount)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "transfer failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

// @Summary Get transactions history for an account
// @Description Retrieve transaction history for a specific bank account
// @ID get-transactions-history
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} AllTransactionsRes
// @Failure 400 {object} ErrorResponse "Invalid account ID format"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/transactions/{id} [get]
// @Security BearerAuth
func (api *ApiManager) handleGetTransactionsHistory(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	transactions, err := api.accMgr.GetTransactionsHistory(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error fetching transactions: %v", err)})
		return
	}

	table, err := utils.FormatTransactionsTable(transactions,idParam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(table))
}

// handleDeposit handles deposit requests
// @Summary Deposit to an account
// @Description Deposit a specified amount to an account
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deposit body DepositRequest true "Deposit Request"
// @Success 200 {object} DepositResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Error depositing to account"
// @Router /account/deposit [post]
func (api *ApiManager) handleDeposit(ctx *gin.Context) {
	var req DepositRequest

	// Parse the JSON request body
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request"})
		return
	}

	// Validate that the amount is positive
	if req.Amount <= 0 {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Amount must be greater than zero"})
		return
	}

	// Convert the account ID from string to ObjectID
	accountID, err := primitive.ObjectIDFromHex(req.AccountID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid account ID format"})
		return
	}

	// Perform the deposit operation
	err = api.accMgr.DepositToAccount(req.Amount, accountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error depositing to account"})
		return
	}

	// Prepare and send the response
	response := DepositResponse{
		Message: "Deposit successful",
		Amount:  req.Amount,
	}
	ctx.JSON(http.StatusOK, response)
}

// handleCheckBalance checks the balance of an account
// @Summary Check account balance
// @Description Check the balance of an account and list recent transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId query string false "Account ID"
// @Param accountHolder query string false "Account Holder Name"
// @Success 200 {object} BalanceResponse
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/balance [get]
func (api *ApiManager) handleCheckBalance(ctx *gin.Context) {
	accountID := ctx.Query("accountId")
	accountName := ctx.Query("accountName")

	if accountID == "" && accountName == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Account ID or name must be provided"})
		return
	}

	// Call handleCheckBalanceIntent
	balance, transactions, err := api.handleCheckBalanceIntent(accountID, accountName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	// Prepare the transaction info
	transactionInfos := []TransactionInfo{}
	for _, transaction := range transactions {
		if transaction.ToAccount.Hex() == accountID {
			transactionInfos = append(transactionInfos, TransactionInfo{
				FromAccount: transaction.FromAccount.Hex(),
				Amount:      transaction.Amount,
			})
		}
	}

	// Prepare and send the response
	response := BalanceResponse{
		Balance:      balance,
		Transactions: transactionInfos,
	}
	ctx.JSON(http.StatusOK, response)
}

func (api *ApiManager) healthCheckHandler(c *gin.Context) {
	response := HealthResponse{Status: "OK"}
	c.JSON(http.StatusOK, response)
}
