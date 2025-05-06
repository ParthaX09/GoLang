package routes

import (
	"database/sql"
	"net/http"
	"server/database"
	"server/middleware"
	"server/models"
	"server/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func UserRoutes(router *gin.Engine) {

	// User registration route
	router.POST("/register", registerHandler)
	router.POST("/login", LoginHandler)

	// Group routes that need authentication
	auth := router.Group("/auth")
	auth.Use(middleware.AuthMiddleware()) // Use AuthMiddleware for authentication
	{
		// Get all users - Protected route
		auth.GET("/users", getAllUsersHandler)

		// Get user by ID - Protected route
		auth.GET("/users/:id", getUserByIdHandler)

		// Update user - Protected route
		auth.PUT("/users/update/:id", updateUserHandler)
	}
}


func registerHandler(c *gin.Context) {
	var newUser models.User

	// Bind request JSON to struct
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if newUser.Name == "" || newUser.Email == "" || newUser.Phone == "" || newUser.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	newUser.Password = hashedPassword

	// Insert into database
	result, err := database.DB.Exec(`
		INSERT INTO users (name, email, phone, password) 
		VALUES (?, ?, ?, ?)`,
		newUser.Name, newUser.Email, newUser.Phone, newUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the inserted user ID
	id, _ := result.LastInsertId()
	newUser.ID = int(id)

	// Respond with success
	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"user":    newUser,
	})
}


func LoginHandler(c *gin.Context) {
	type LoginInput struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch user from DB
	var user models.User
	err := database.DB.QueryRow("SELECT id, password FROM users WHERE email = ?", input.Email).
		Scan(&user.ID, &user.Password)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}

// Get all users handler (protected)
func getAllUsersHandler(c *gin.Context) {
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
}

// Get user by ID handler (protected)
func getUserByIdHandler(c *gin.Context) {
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
}

// Update user handler (protected)
func updateUserHandler(c *gin.Context) {
	id := c.Param("id")

	// Fetch existing user
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

	// Binding input fields
	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
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
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		existingUser.Password = hashedPassword
	}

	// Update in the database
	_, err = database.DB.Exec(`
		UPDATE users 
		SET name = ?, email = ?, phone = ?, password = ? 
		WHERE id = ?`,
		existingUser.Name, existingUser.Email, existingUser.Phone,
		existingUser.Password, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the updated user details
	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    existingUser,
	})
}
	
	// // Create user
	// router.POST("/register", func(c *gin.Context) {
	// 	var newUser models.User

	// 	// Bind the request body to newUser
	// 	if err := c.ShouldBindJSON(&newUser); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	// Hash the password
	// 	hashedPassword, err := utils.HashPassword(newUser.Password)
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
	// 		return
	// 	}
	// 	newUser.Password = hashedPassword
	
	// 	// Insert into DB with all relevant fields
	// 	result, err := database.DB.Exec(`
	// 		INSERT INTO users (name, email, phone, password) 
	// 		VALUES (?, ?, ?, ?)`,
	// 		newUser.Name, newUser.Email, newUser.Phone, newUser.Password)
	
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}

	// 	// Get the inserted user ID
	// 	id, _ := result.LastInsertId()
	// 	newUser.ID = int(id)

	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "User registered successfully",
	// 		"user":    newUser,
	// 	})
	// })

// 	// Get all users
// 	router.GET("/users", func(c *gin.Context) {
// 		rows, err := database.DB.Query("SELECT id, name, email, phone, password, created, updated FROM users")
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		defer rows.Close()

// 		var users []models.User
// 		for rows.Next() {
// 			var user models.User
// 			if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.Created, &user.Updated); err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 				return
// 			}
// 			users = append(users, user)
// 		}
// 		c.JSON(http.StatusOK, users)
// 	})

// 	// Get user by ID
// 	router.GET("/users/:id", func(c *gin.Context) {
// 		id := c.Param("id")
// 		var user models.User
// 		err := database.DB.QueryRow("SELECT id, name, email, phone, password, created, updated FROM users WHERE id = ?", id).
// 			Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.Created, &user.Updated)

// 		if err == sql.ErrNoRows {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 			return
// 		} else if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}
// 		c.JSON(http.StatusOK, user)
// 	})

	
// 	//Update the user
// 	router.PUT("/users/update/:id", func(c *gin.Context) {
// 		id := c.Param("id")

// 		// Fetching existing user
// 		var existingUser models.User
// 		err := database.DB.QueryRow("SELECT id, name, email, phone, password, created, updated FROM users WHERE id = ?", id).
// 			Scan(&existingUser.ID, &existingUser.Name, &existingUser.Email, &existingUser.Phone, &existingUser.Password, &existingUser.Created, &existingUser.Updated)
// 		if err == sql.ErrNoRows {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 			return
// 		} else if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		var input map[string]interface{}
// 		if err := c.ShouldBindJSON(&input); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		//Update only provided fields
// 		if name, ok := input["name"].(string); ok {
// 			existingUser.Name = name
// 		}
// 		if email, ok := input["email"].(string); ok {
// 			existingUser.Email = email
// 		}
// 		if phone, ok := input["phone"].(string); ok {
// 			existingUser.Phone = phone
// 		}
// 		if password, ok := input["password"].(string); ok {
// 			hashedPassword, err := utils.HashPassword(password)
// 			if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
// 			return
// 		}
// 			existingUser.Password = hashedPassword
// 		}

// 		_, err = database.DB.Exec(`
// 			UPDATE users 
// 			SET name = ?, email = ?, phone = ?, password = ?
// 			WHERE id = ?`,
// 			existingUser.Name, existingUser.Email, existingUser.Phone,
// 			existingUser.Password, id)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "User updated successfully",
// 			"user":    existingUser,
// 		})
// 	})
//}
