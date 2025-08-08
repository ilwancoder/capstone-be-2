package routes

import (
	"os"

	"capstone-be-2/controllers"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

// BookRoutes configures routes for the "books" API
func BookRoutes(app *fiber.App) {
	// Group for public (unprotected) routes
	public := app.Group("/api")

	// Public routes
	public.Get("/books", controllers.GetAllBooks)
	public.Get("/books/:id", controllers.GetBookByID)

	// Get JWT_SECRET from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")

	// Create a new API group for protected routes
	api := app.Group("/api")
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(jwtSecret)},
	}))

	// Protected routes
	api.Post("/books", controllers.CreateBook)
	api.Post("/books/:id/borrow", controllers.BorrowBook)
	api.Post("/books/:id/return", controllers.ReturnBook)

}
