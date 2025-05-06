package main

import (
	"log"
	"server/database"
	"server/routes"
	"github.com/gin-gonic/gin"
)

func main() {

	if err := database.Connect(); err != nil {
        log.Fatal(err)
    }
	
	router := gin.Default()
	routes.UserRoutes(router)
	router.Run(":8080")
}
