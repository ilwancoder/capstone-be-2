package controllers

import (
	"context"
	"strings"
	"time"

	"capstone-be-2/config"
	"capstone-be-2/controllers/services"
	"capstone-be-2/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterUser(c *fiber.Ctx) error {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid request body"})
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	if body.Email == "" || body.Password == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Email & password required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	users := config.GetCollection("users")

	// email unik
	if err := users.FindOne(ctx, bson.M{"email": body.Email}).Err(); err == nil {
		return c.Status(409).JSON(fiber.Map{"message": "Email already registered"})
	}

	hash, err := services.HashPassword(body.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Could not hash password"})
	}

	user := models.User{
		ID:            primitive.NewObjectID(),
		Name:          body.Name,
		Email:         body.Email,
		PasswordHash:  hash,
		Points:        0,
		BorrowedBooks: []primitive.ObjectID{},
	}
	if _, err := users.InsertOne(ctx, user); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Could not register user"})
	}
	return c.Status(201).JSON(fiber.Map{"message": "User registered successfully"})
}

func LoginUser(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid credentials"})
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	users := config.GetCollection("users")

	var user models.User
	if err := users.FindOne(ctx, bson.M{"email": body.Email}).Decode(&user); err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Invalid email or password"})
	}
	if err := services.ComparePassword(user.PasswordHash, body.Password); err != nil {
		return c.Status(401).JSON(fiber.Map{"message": "Invalid email or password"})
	}

	token, err := services.GenerateJWT(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Could not generate token"})
	}
	return c.JSON(fiber.Map{"token": token})
}

func GetCurrentUser(c *fiber.Ctx) error {
	tok := c.Locals("user").(*jwt.Token)
	claims := tok.Claims.(jwt.MapClaims)
	idHex, ok := claims["id"].(string)
	if !ok || idHex == "" {
		return c.Status(401).JSON(fiber.Map{"message": "Invalid token claims"})
	}
	objID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid user id"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	users := config.GetCollection("users")

	var user models.User
	if err := users.FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "User not found"})
	}
	user.PasswordHash = "" // jangan bocorkan hash
	return c.JSON(user)
}
