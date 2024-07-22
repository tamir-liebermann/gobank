package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tamir-liebermann/gobank/api"
	"github.com/tamir-liebermann/gobank/db"
	_ "github.com/tamir-liebermann/gobank/docs"
)

// indexHandler responds to requests with our greeting.
// @title gobank
// @version 1.0
// @description This is the main function that initializes the database, API manager, and starts the Gin server.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /api/v1

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Initialize database, API manager, and Gin router
	accMgr, err := db.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	apiMgr := api.NewApiManager(accMgr)
	router := gin.Default()
	apiMgr.RegisterRoutes(router)
  
	// Start the server
	apiMgr.Run()
}
