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
	// Create user
	// r.POST("/register", func(c *gin.Context) {
	// 	var newUser models.User
	// 	if err := c.ShouldBindJSON(&newUser); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	result, err := database.DB.Exec("INSERT INTO users (name, email) VALUES (?, ?)", newUser.Name, newUser.Email)
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// 	id, _ := result.LastInsertId()
	// 	newUser.ID = int(id)

	// 	c.JSON(http.StatusOK, gin.H{"message": "User registered", "user": newUser})
	// })

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
		newUser.Updated = now

		// Insert into DB with all relevant fields
		result, err := database.DB.Exec(`
			INSERT INTO users (name, email, phone, password, created, updated) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			newUser.Name, newUser.Email, newUser.Phone, newUser.Password, newUser.Created, newUser.Updated)

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
		rows, err := database.DB.Query("SELECT id, name, email, phone, password, created, updated FROM users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.Created, &user.Updated); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			users = append(users, user)
		}
		c.JSON(http.StatusOK, users)
	})

	// Get user by ID
	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		err := database.DB.QueryRow("SELECT id, name, email, phone, password, created, updated FROM users WHERE id = ?", id).
			Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.Created, &user.Updated)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	// //Update the user
	// r.PUT("/users/update/:id",func(c *gin.Context) {
	// 	id:= c.Param("id")
	// 	var updatedUser models.User
	// 	if err := c.ShouldBindJSON(&updatedUser); err != nil{
	// 		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
	// 		return
	// 	}

	// 	//update th timestamp
	// 	updatedUser.Updated = time.Now()

	// 	//update query
	// 	result,err := database.DB.Exec(`UPDATE users SET name = ?, email = ?, phone = ?, password = ?, updated = ? WHERE id = ?`, updatedUser.Name,updatedUser.Email,updatedUser.Phone,updatedUser.Password,updatedUser.Updated,id)
	// 	if err != nil{
	// 		c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	// 		return
	// 	}

	// 	//Check if any row was updated actually
	// 	rowsAffected, err := result.RowsAffected()
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// 	if rowsAffected == 0 {
	// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	// 		return
	// 	}
	// 	updatedUser.ID = func() int {
	// 		var i int
	// 		fmt.Sscanf(id, "%d" ,&i)
	// 		return i
	// 	}()

	// 	c.JSON(http.StatusOK,gin.H{
	// 		"message":"User Updated Successfully",
	// 		"user":updatedUser,
	// 	})
	// })

	//Update the user
	r.PUT("/users/update/:id", func(c *gin.Context) {
		id := c.Param("id")

		// Fetching existing user
		var existingUser models.User
		err := database.DB.QueryRow("SELECT id, name, email, phone, password, created, updated FROM users WHERE id = ?", id).
			Scan(&existingUser.ID, &existingUser.Name, &existingUser.Email, &existingUser.Phone, &existingUser.Password, &existingUser.Created, &existingUser.Updated)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var input map[string]interface{}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//Update only provided fields
		if name, ok := input["name"].(string); ok {
			existingUser.Name = name
		}
		if email, ok := input["email"].(string); ok {
			existingUser.Email = email
		}
		if phone, ok := input["phone"].(string); ok {
			existingUser.Phone = phone
		}
		if password, ok := input["password"].(string); ok {
			existingUser.Password = password
		}

		// Update the timestamp
		existingUser.Updated = time.Now()

		_, err = database.DB.Exec(`
			UPDATE users 
			SET name = ?, email = ?, phone = ?, password = ?, updated = ? 
			WHERE id = ?`,
			existingUser.Name, existingUser.Email, existingUser.Phone,
			existingUser.Password, existingUser.Updated, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User updated successfully",
			"user":    existingUser,
		})
	})

}
