package routes

import (
	"capstone-be-2/controllers"

	"github.com/gofiber/fiber/v2"
)

// AuthRoutes handles all authentication-related routes
func AuthRoutes(app *fiber.App) {
	auth := app.Group("/api/auth")

	// Define auth routes
	auth.Post("/register", controllers.RegisterUser)
	auth.Post("/login", controllers.LoginUser)
}
