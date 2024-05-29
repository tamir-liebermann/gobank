package api

import (
	// "log"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

	server.POST("/create", api.handleCreateAccount)
	router := server.Group("/")
	router.Use(authenticate) 
	router.POST("/login", api.handleLogin)

	accounts := server.Group("/account")
	accounts.Use(authenticate)

	accounts.GET("/:id", api.handleGetById)
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


func(api *ApiManager)  handleCreateAccount(ctx *gin.Context) {
	
	var req CreateAccountRequest
	err := ctx.ShouldBindJSON(&req)
	if err !=nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "bad reuqest"})
		
	}
	
	 accountID,err := api.accMgr.CreateAccount(req.UserName,req.Password) 
	if err != nil {
		
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create account. Try again later."})
		return
	}	

	objectID, err := primitive.ObjectIDFromHex(accountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating account ID"})
		return
	}
	token, err := utils.GenerateToken(req.UserName, objectID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating token"})
		return
	}
	response := CreateAccountResponse{ 
			Message: "Account created!",
			Id:      accountID,
			Token:   token,
	}

	
	ctx.JSON(http.StatusCreated, response)
}

func (api *ApiManager) handleLogin(ctx *gin.Context) {
    var req LoginRequest

    err := ctx.ShouldBindJSON(&req)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "Bad Request", "error": err.Error()})
        return
    }

    if req.UserName == "" || req.Password == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "Username and password are required"})
        return
    }

    account, err := api.accMgr.SearchAccountByName(req.UserName)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
        return
    }

    if account == nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
        return
    }

    if !utils.CheckPasswordHash(req.Password, account.Password) {
        ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
        return
    }

    token, err := utils.GenerateToken(account.AccountHolder, account.ID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Could not generate token"})
        return
    }
    response := LoginResponse{
        UserName: req.UserName,
        Token:    token,
    }

    ctx.JSON(http.StatusOK, response)
	
}

func (api *ApiManager) handleGetById(ctx *gin.Context) {
    idParam := ctx.Param("id")
    id, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
        return
    }

    account, err := api.accMgr.SearchAccountById(id)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
        return
    }

    if account == nil {
        ctx.JSON(http.StatusNotFound, gin.H{"message": "Account not found"})
        return
    }
	ctx.JSON(http.StatusOK, account)

}

func (api *ApiManager) handleDeleteById(ctx *gin.Context) {
	idParam := ctx.Param("id")
    id, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID format"})
        return
    }
   err = api.accMgr.DeleteAccountById(id)
   if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
        return
   }

	ctx.JSON(http.StatusOK, gin.H{"message": "Account Deleted Successfully!"})
}

func (api *ApiManager) handleGetAccounts(ctx *gin.Context) {
	 accounts, err := api.accMgr.GetAccounts()
	 if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve accounts"})
        return
	 }

	 ctx.JSON(http.StatusOK, accounts)
}

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







	
