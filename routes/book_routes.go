package routes

import (
	"os"

	"capstone-be-2/controllers"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func BookRoutes(app *fiber.App) {
	// PUBLIC
	public := app.Group("/api")
	public.Get("/books", controllers.GetAllBooks)
	public.Get("/books/:id", controllers.GetBookByID)
	public.Get("/leaderboard", controllers.GetLeaderboard)

	// PROTECTED
	protected := app.Group("/api", jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		AuthScheme: "Bearer",
	}))
	protected.Post("/books", controllers.CreateBook) // optional (admin seeding)
	protected.Post("/books/:id/borrow", controllers.BorrowBook)
	protected.Post("/books/:id/return", controllers.ReturnBook)
}
