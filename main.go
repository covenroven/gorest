package main

import (
	"github.com/covenroven/gorest/database"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		panic(err)
	}
}
