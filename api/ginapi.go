package api

import (
	// "log"

	"net/http"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tamir-liebermann/gobank/db"
	"github.com/tamir-liebermann/gobank/utils"
)

type ApiManager struct {
	accMgr *db.AccManager
}

func NewApiManager(mgr *db.AccManager) *ApiManager {
	return &ApiManager{
		accMgr: mgr,
	}
}

func (api *ApiManager) registerRoutes(server *gin.Engine) {
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	server.POST("/create", api.handleCreateAccount)
	server.POST("/login", api.handleLogin)

	accounts := server.Group("/account")
	accounts.Use(authenticate)

	accounts.GET("/transactions/:id", api.handleGetTransactionsHistory)
	accounts.GET("/:id", api.handleGetById)
	accounts.GET("/name/:account_holder", api.handleGetByNameOrPhone)
	accounts.DELETE("/:id", api.handleDeleteById)
	accounts.POST("/transfer/:id", api.handleTransfer)
	accounts.POST("/chatgpt", api.handleChatGPTRequest)
	accounts.POST("/deposit", api.handleDeposit)

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

func (api *ApiManager) Run() {

	server := gin.Default()
	api.registerRoutes(server)

	server.Run(":8080") //localhost:8080
}
