package api

import (
	// "log"

	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tamir-liebermann/gobank/db"
	"github.com/tamir-liebermann/gobank/env"
	"github.com/tamir-liebermann/gobank/utils"
	"github.com/twilio/twilio-go"
)




type ApiManager struct {
	accMgr        *db.AccManager
	twilioClient  *twilio.RestClient
}

func NewApiManager(mgr *db.AccManager) *ApiManager {
	 spec := env.New()


	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: spec.TwilioAccSid,
		Password: spec.TwilioAuth,
	})

	return &ApiManager{
		accMgr:       mgr,
		twilioClient: twilioClient,
	}
}

func (api *ApiManager) RegisterRoutes(server *gin.Engine) {
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	server.POST("/create", api.handleCreateAccount)
	server.POST("/login", api.handleLogin)

	accounts := server.Group("/account")
	accounts.Use(authenticate)

	accounts.GET("/transactions/:id", api.handleGetTransactionsHistory)
	accounts.GET("/:id", api.handleGetById)
	accounts.GET("/balance", api.handleCheckBalance)
	accounts.GET("/name/:account_holder", api.handleGetByNameOrPhone)
	accounts.DELETE("/:id", api.handleDeleteById)
	accounts.POST("/transfer", api.handleTransfer)
	accounts.POST("/chatgpt", api.handleChatGPTRequest)
	accounts.POST("/deposit", api.handleDeposit)
	

	admin := server.Group("/admin")
	admin.GET("/accounts", api.handleGetAccounts)
	server.POST("/webhook", api.handleTwilioWebhook)
	server.GET("/health", api.healthCheckHandler)
	

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
	api.RegisterRoutes(server)

	port := os.Getenv("PORT")
    if port == "" {
        port = "5252"
    }

    server.Run(":" + port)
}

