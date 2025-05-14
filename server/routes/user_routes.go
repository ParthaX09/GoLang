package routes

import (
	"server/handler"
	"server/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {

	// User registration route
	router.POST("/register", handler.RegisterHandler)
	router.POST("/login", handler.LoginHandler)

	// Group routes that need authentication
	auth := router.Group("/auth")
	auth.Use(middleware.AuthMiddleware())

	// Use AuthMiddleware for authentication
	{
		// Get all users - Protected route
		auth.GET("/users", middleware.RoleMiddleware("admin", "sub_admin", "client"), handler.GetAllUsersHandler)

		// Get user by ID - Protected route
		auth.GET("/users/:id", middleware.RoleMiddleware("admin", "sub_admin", "client"), handler.GetUserByIdHandler)

		// Update user - Protected route
		auth.PUT("/users/update/:id", middleware.RoleMiddleware("admin"), handler.UpdateUserHandler)
	}
}
