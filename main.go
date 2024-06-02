package main

import (
	"github.com/tamir-liebermann/gobank/api"
	"github.com/tamir-liebermann/gobank/db"
	_ "github.com/tamir-liebermann/gobank/docs"
)

// @title GoBank API
// @version 1.0
// @description This is a sample server for a banking application.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /
func main() {


	accMgr,err := db.InitDB()
	if err !=nil{
		panic(err)
	}

	apiMgr := api.NewApiManager(accMgr)
	apiMgr.Run()
}
