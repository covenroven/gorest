package main

import (
	"github.com/covenroven/gorest/api"
	"github.com/covenroven/gorest/config"
	"github.com/covenroven/gorest/database"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	app := &api.App{
		DB: db,
	}

	r := router(app)
	r.Run(":" + config.SRV_PORT)
}

// Defines route for the service
func router(app *api.App) *gin.Engine {
	r := gin.Default()

	orders := r.Group("orders")
	orders.GET("", app.GetOrders)
	orders.GET(":orderID", app.ShowOrder)
	orders.POST("", app.CreateOrder)
	orders.PUT(":orderID", app.UpdateOrder)
	orders.DELETE(":orderID", app.DeleteOrder)

	return r
}
