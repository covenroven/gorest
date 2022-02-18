package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Main api application that contains database struct
type App struct {
	DB *sql.DB
}

// Return internal server error response and print the error
func throwServerError(c *gin.Context, err error) {
	fmt.Println("[Error]", err)
	res := gin.H{
		"errors": "Internal server error",
	}
	c.JSON(http.StatusInternalServerError, res)
}
