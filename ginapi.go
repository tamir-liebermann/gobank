package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tamir-liebermann/gobank/utils"
)

func registerRoutes(server *gin.Engine) {


router := server.Group("/")
// router.Use(middlewares.Authenticate) // todo implement this later 
router.POST("/login", handleLogin)
router.POST("/account", handleCreateAccount)

accounts := server.Group("/account")
accounts.Use(authenticate)

accounts.GET("/:id", handleGetById)
accounts.DELETE("/:id", handleDeleteById)
accounts.POST("/transfer/:id", handleTransfer )


admin := server.Group("/admin")
accounts.Use(authenticate)

admin.GET("/accounts", handleGetAccounts)

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
func Run(){
	server := gin.Default() 
	registerRoutes(server)


	server.Run(":8080") //localhost:8080
}