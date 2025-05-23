package handler

import (
	"net/http"
	"server/database"
	"server/models"
	"server/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(c *gin.Context) {
	var newUser models.User

	// Bind request JSON to struct
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if newUser.Name == "" || newUser.Email == "" || newUser.Phone == "" || newUser.Password == "" || newUser.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":     true,
		"sub_admin": true,
		"client":    true,
	}
	if !validRoles[newUser.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'admin', 'sub_admin', or 'client'"})
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
	id, err := database.CreateUser(&newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
	user, err := database.GetUserByEmail(input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID, user.Role)
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
func GetAllUsersHandler(c *gin.Context) {
	role := c.MustGet("userRole").(string)

	//Fetch from DB
	rows, err := database.GetAllUsersByRole(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.Role, &user.Created, &user.Updated); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}


// Get user by ID handler (protected)
func GetUserByIdHandler(c *gin.Context) {
	id := c.Param("id")
	requestorRole := c.MustGet("userRole").(string)
	requestorID := c.MustGet("userID").(int)

	// Fetch the target user
	targetUser, err := database.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Apply RBAC
	switch requestorRole {
	case "admin":
		// Admin can access any user — allow
	case "sub_admin":
		if targetUser.Role == "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Sub-admins cannot access admin data"})
			return
		}
	case "client":
		if targetUser.Role != "client" && requestorID != targetUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Clients can only access other clients or their own data"})
		return
	}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized role"})
		return
	}

	c.JSON(http.StatusOK, targetUser)
}


// Update user handler (protected)
func UpdateUserHandler(c *gin.Context) {
	id := c.Param("id")

	// Fetch existing user
	existingUser, err := database.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
	if role, ok := input["role"].(string); ok {
		existingUser.Role = role
	}

	// Update in the database
	if err := database.UpdateUser(existingUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the updated user details
	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    existingUser,
	})
}