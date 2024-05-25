package main

import (
	"github.com/tamir-liebermann/gobank/api"
	"github.com/tamir-liebermann/gobank/db"
)


func main() {
	accMgr,err := db.InitDB()
	if err !=nil{
		panic(err)
	}

	apiMgr := api.NewApiManager(accMgr)
	apiMgr.Run()
}
