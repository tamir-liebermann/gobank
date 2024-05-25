package api

import (
	// "log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tamir-liebermann/gobank/db"
	"github.com/tamir-liebermann/gobank/utils"
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

	router := server.Group("/")
	// router.Use(middlewares.Authenticate) // todo implement this later
	// router.POST("/login", handleLogin)
	router.POST("/create", api.handleCreateAccount)

	accounts := server.Group("/account")
	accounts.Use(authenticate)

	// accounts.GET("/:id", handleGetById)
	// accounts.DELETE("/:id", handleDeleteById)
	// accounts.POST("/transfer/:id", handleTransfer )

	// admin := server.Group("/admin")
	// accounts.Use(authenticate)

	// admin.GET("/accounts", handleGetAccounts)

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

	ctx.JSON(http.StatusCreated, gin.H{"message": "account created!", "id": accountID})
}
