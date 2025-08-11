package main

import (
	"capstone-be-2/config"
	"capstone-be-2/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load file .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	// Ambil variabel dari environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set in .env")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in .env")
	}

	app := fiber.New()
	log.Println("Starting server with MongoDB URI:", mongoURI)

	// Inisialisasi koneksi MongoDB
	config.ConnectDB()

	// Aktivasi CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// --- Lakukan semua inisialisasi rute di sini ---

	// Setup Auth Routes (Unprotected)
	routes.AuthRoutes(app)

	// Setup Book Routes (Protected)
	routes.BookRoutes(app)

	// Health check route
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "OK",
		})
	})

	// --- Panggil app.Listen() di bagian paling akhir ---
	// Server mulai mendengarkan permintaan
	log.Fatal(app.Listen(":8000"))
}
