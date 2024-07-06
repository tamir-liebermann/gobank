package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tamir-liebermann/gobank/api"
	"github.com/tamir-liebermann/gobank/db"
)






// indexHandler responds to requests with our greeting.

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
