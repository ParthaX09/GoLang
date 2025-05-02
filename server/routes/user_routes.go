package routes

import (
	"database/sql"
	"net/http"
	"server/database"
	"server/models"
	"time"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	
	//create user
	r.POST("/register", func(c *gin.Context) {
		var newUser models.User
	
		// Bind the request body to newUser
		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	
		// Manually set Created and Updated timestamps
		now := time.Now()
		newUser.Created = now
		
	
		// Insert into DB with all relevant fields
		result, err := database.DB.Exec(`
			INSERT INTO users (name, email, phone, password, created, updated) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			newUser.Name, newUser.Email, newUser.Phone, newUser.Password, newUser.Created, nil)
	
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	
		// Get the inserted user ID
		id, _ := result.LastInsertId()
		newUser.ID = int(id)
	
		c.JSON(http.StatusOK, gin.H{
			"message": "User registered successfully",
			"user":    newUser,
		})
	})
	

	// Get all users
	r.GET("/users", func(c *gin.Context) {
		rows, err := database.DB.Query("SELECT id, name, email FROM users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			users = append(users, u)
		}
		c.JSON(http.StatusOK, users)
	})

	// Get user by ID
	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		err := database.DB.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).
			Scan(&user.ID, &user.Name, &user.Email)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})
}
