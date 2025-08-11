package routes

import (
	"os"

	"capstone-be-2/controllers"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	// PUBLIC
	auth := app.Group("/api/auth")
	auth.Post("/register", controllers.RegisterUser)
	auth.Post("/login", controllers.LoginUser)
	auth.Get("/ping", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"ok": true}) })

	// PROTECTED
	protected := app.Group("/api/auth", jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		AuthScheme: "Bearer",
	}))
	protected.Get("/me", controllers.GetCurrentUser)
	protected.Get("/users/me/borrowed", controllers.GetMyBorrowed) // ← tambah ini

}
