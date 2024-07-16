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
	accMgr       *db.AccManager
	twilioClient *twilio.RestClient
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
	server.POST("/webhook", api.twilioAuthenticate, api.handleTwilioWebhook, )
	server.GET("/health", api.healthCheckHandler)

}


func (api *ApiManager) twilioAuthenticate(ctx *gin.Context) {
    var twilioReq TwilioReq
	spec := env.New()
	if err := ctx.ShouldBind(&twilioReq); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	bodyWithoutSecret, secret := extractSecretFromBody(twilioReq.Body)
	if secret == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Missing Twilio secret"})
		return
	}
	twilioReq.Body = bodyWithoutSecret

	// Handle Twilio request authentication
	if secret != spec.TwilioSecret {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid Twilio secret"})
		return
	}
	
    // Bind the request body to the TwilioReq struct

 // Load environment variables or configuration

    // Validate the secret
    
    // Handle Twilio request authentication
    account, err := api.getAccountFromTwilioReq(ctx, twilioReq)
    if err != nil {
        ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
        return
    }

    ctx.Set("userId", account.ID.Hex())
    ctx.Set("twilioSecret", twilioReq.Secret)

    ctx.Next()
}
   
   



func authenticate(context *gin.Context) {
// first, if this is twilio req implement getAccountFromTwilioReq here...
	
	
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
