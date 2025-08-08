package controllers

import (
	"context"
	"strings"

	"capstone-be-2/config"
	"capstone-be-2/controllers/services"
	"capstone-be-2/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RegisterUser handles new user registration
func RegisterUser(c *fiber.Ctx) error {
	// ... (logic for parsing body and checking email)
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	user.Email = strings.ToLower(user.Email)
	collection := config.GetCollection("users")
	if err := collection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&user); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Email already registered"})
	}

	// Call service to hash password
	hashedPassword, err := services.HashPassword(user.PasswordHash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Could not hash password"})
	}
	user.PasswordHash = hashedPassword

	// ... (logic for setting initial values and inserting into DB)
	user.ID = primitive.NewObjectID()
	user.Points = 0
	user.BorrowedBooks = []primitive.ObjectID{}

	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Could not register user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

// LoginUser handles user login and JWT generation
func LoginUser(c *fiber.Ctx) error {
	// ... (logic for parsing body and finding user)
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&credentials); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid credentials"})
	}
	credentials.Email = strings.ToLower(credentials.Email)

	collection := config.GetCollection("users")
	var user models.User
	if err := collection.FindOne(context.Background(), bson.M{"email": credentials.Email}).Decode(&user); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid email or password"})
	}

	// Call service to compare password
	if err := services.ComparePassword(user.PasswordHash, credentials.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid email or password"})
	}

	// Call service to generate JWT
	token, err := services.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Could not generate token"})
	}

	return c.JSON(fiber.Map{"message": "Login successful", "token": token})
}

func BorrowBook(c *fiber.Ctx) error {
	// Ambil data token dari context
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)

	// Dapatkan user ID dari klaim token
	userID, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token claims"})
	}

	// Konversi user ID string menjadi ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID in token"})
	}

	// Gunakan objID untuk operasi database (misalnya, memperbarui dokumen user)
	// ...

	return c.SendString("This is a protected route. User ID: " + objID.Hex())
}
