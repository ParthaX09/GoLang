package routes

import (
	"net/http"
	"server/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	r.POST("/register", func(c *gin.Context) {
		var newUser models.User
		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newUser.ID = getNextUserID()
		models.Users = append(models.Users, newUser)

		// models.Users = append(models.Users, newUser)
		c.JSON(http.StatusOK, gin.H{
			"message": "User registered successfully",
			"user":    newUser,
		})
	})

	r.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.Users)
	})


	r.GET("/users/:id", func(c *gin.Context) {
		id_param := c.Param("id")
		id, err := strconv.Atoi(id_param)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid ID"})
			return
		}
		for _, user := range models.Users{
			if user.ID == id {
				c.JSON(http.StatusOK, user)
			}
		}
	})
}



func getNextUserID() int {
	if len(models.Users) == 0 {
		return 1
	}
	lastUser := models.Users[len(models.Users)-1]
	return lastUser.ID + 1
}