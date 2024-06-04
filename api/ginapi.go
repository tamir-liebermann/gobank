package api

import (
	// "log"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tamir-liebermann/gobank/db"
	"github.com/tamir-liebermann/gobank/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApiManager struct{
accMgr *db.AccManager
}

func NewApiManager(mgr *db.AccManager)*ApiManager{
return &ApiManager{
	accMgr: mgr,
}
}

func(api *ApiManager) registerRoutes(server *gin.Engine) {
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	server.POST("/create", api.handleCreateAccount)
	router := server.Group("/")
	router.Use(authenticate) 
	router.POST("/login", api.handleLogin)

	accounts := server.Group("/account")
	accounts.Use(authenticate)

	accounts.GET("/transactions/:id", api.handleGetTransactionsHistory)
	accounts.GET("/:id", api.handleGetById)
    accounts.GET("/name/:account_holder", api.handleGetByName)
	accounts.DELETE("/:id", api.handleDeleteById)
	accounts.POST("/transfer/:id", api.handleTransfer )

	admin := server.Group("/admin")
	admin.GET("/accounts", api.handleGetAccounts)

}

func authenticate(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")

	if token == "" {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized"})
		return
	}

	userId, err := utils.VerifyToken(token)

	if err != nil {
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized"})
		return
	}

	context.Set("userId", userId)
	context.Next()
}


func(api *ApiManager) Run() {
	
	server := gin.Default()
	api.registerRoutes(server)

	server.Run(":8080") //localhost:8080
}

// @Summary Create a new account
// @Description Create a new account with username and password
// @ID create-account
// @Accept  json
// @Produce  json
// @Param   account  body     CreateAccountRequest  true  "Account Information"
// @Success 201 {object} string "Account created!"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /create [post]
func(api *ApiManager)  handleCreateAccount(ctx *gin.Context) {
	
	var req CreateAccountRequest
	err := ctx.ShouldBindJSON(&req)
	if err !=nil{
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Bad request"})
		
	}
	
	 accountID,err := api.accMgr.CreateAccount(req.UserName,req.Password) 
	if err != nil {
		
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Could not create account, try again later"})
		return
	}	

	objectID, err := primitive.ObjectIDFromHex(accountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generating account id "})
		return
	}
	token, err := utils.GenerateToken(req.UserName, objectID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generating token "})
		return
	}
	response := CreateAccountResponse{ 
			Message: "Account created!",
			Id:      accountID,
			Token:   token,
	}

	
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
// @Security BearerAuth
func (api *ApiManager) handleLogin(ctx *gin.Context) {
    var req LoginRequest

    err := ctx.ShouldBindJSON(&req)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Bad request"})
        return
    }

    if req.UserName == "" || req.Password == "" {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Username and password are required"})
        return
    }

    account, err := api.accMgr.SearchAccountByName(req.UserName)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error"})
        return
    }

    if account == nil {
        ctx.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid cardinals"})
        return
    }

    if !utils.CheckPasswordHash(req.Password, account.Password) {
        ctx.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Invalid cardinals"})
        return
    }

    token, err := utils.GenerateToken(account.AccountHolder, account.ID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generating token "})
        return
    }
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
// @ID get-account-by-name
// @Produce json
// @Param account_holder path string true "Account Holder's Name"
// @Success 200 {object} BankAccRes    "Account found!"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Account not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/name/{account_holder} [get]
// @Security BearerAuth
func (api *ApiManager) handleGetByName(ctx *gin.Context) {
    // Extract the account holder's name from the request parameters
    accountHolder := ctx.Param("account_holder")

    // Search for the account by name
    account, err := api.accMgr.SearchAccountByName(accountHolder)
    if err != nil {
        // If there's an internal server error, return 500 status
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error"})
        return
    }

    // If the account is not found, return 404 status
    if account == nil {
        ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "Account not found"})
        return
    }

    // Return the account details if found
    ctx.JSON(http.StatusOK, account)
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
	 accounts, err := api.accMgr.GetAccounts()
	 if err != nil{
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
// @Param id path string true "Account ID"
// @Success 200 {object} BankAccRes 
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Invalid account ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/transfer/{id} [post]
// @Security BearerAuth
func (api *ApiManager) handleTransfer(ctx *gin.Context) {
    var req TransferRequest

    // Use ShouldBindJSON to parse the request body
    if err := ctx.ShouldBindJSON(&req); err != nil {
        log.Printf("Error binding JSON: %v", err)
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
        return
    }

    fromAccountID, err := primitive.ObjectIDFromHex(req.FromAccountID)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid from account ID"})
        return
    }

    toAccountID, err := primitive.ObjectIDFromHex(req.ToAccountID)
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
// @Success 200 {array} TransactionRes "Successful response"
// @Failure 400 {object} ErrorResponse "Invalid account ID format"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/transactions/{id} [get]
// @Security BearerAuth
// handleGetTransactionsHistory handles the GET request to retrieve transaction history for a specific account.
func (api *ApiManager) handleGetTransactionsHistory(ctx *gin.Context) {
    idParam := ctx.Param("id")
    id, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid ID format"})
        return
    }

    transactions, err := api.accMgr.GetTransactionsHistory(id)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Internal server error"})
        return
    }

    if transactions == nil {
        ctx.JSON(http.StatusNotFound, ErrorResponse{Message: "Transactions not found"})
        return
    }

    ctx.JSON(http.StatusOK, transactions)
}







	
