package controllers

import (
	"context"
	"time"

	"capstone-be-2/config"
	"capstone-be-2/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetAllBooks handles GET /api/books to retrieve all books
func GetAllBooks(c *fiber.Ctx) error {
	collection := config.GetCollection("books")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var books []models.Book
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching books",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &books); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error decoding books",
			"error":   err.Error(),
		})
	}

	return c.JSON(books)
}

// GetBookByID handles GET /api/books/:id to retrieve a single book
func GetBookByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid book ID format",
		})
	}

	collection := config.GetCollection("books")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var book models.Book
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&book)
	if err != nil {
		// If book is not found, return 404
		if err.Error() == "mongo: no documents in result" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Book not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching book",
			"error":   err.Error(),
		})
	}

	return c.JSON(book)
}

func BorrowwBook(c *fiber.Ctx) error {
	// 1. Extract user ID from JWT token claims
	userToken := c.Locals("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userID, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token claims"})
	}

	objUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user ID in token"})
	}

	// 2. Parse book ID from URL parameter
	bookID := c.Params("id")
	objBookID, err := primitive.ObjectIDFromHex(bookID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid book ID format"})
	}

	// 3. Find book and check availability
	booksCollection := config.GetCollection("books")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var book models.Book
	err = booksCollection.FindOne(ctx, bson.M{"_id": objBookID}).Decode(&book)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Book not found"})
	}

	if book.Available <= 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Book is currently not available"})
	}

	// 4. Find user and check if the book is already borrowed
	usersCollection := config.GetCollection("users")
	var user models.User
	err = usersCollection.FindOne(ctx, bson.M{"_id": objUserID}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "User not found"})
	}

	for _, borrowedBookID := range user.BorrowedBooks {
		if borrowedBookID == objBookID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Book already borrowed by user"})
		}
	}

	// 5. Update user and book records
	// Atomically add the book to the user's BorrowedBooks list
	updateUser := bson.M{
		"$push": bson.M{"borrowed_books": objBookID},
	}
	_, err = usersCollection.UpdateOne(ctx, bson.M{"_id": objUserID}, updateUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update user's borrowed books"})
	}

	// Atomically decrement the book's availability count
	updateBook := bson.M{
		"$inc": bson.M{"available": -1},
	}
	_, err = booksCollection.UpdateOne(ctx, bson.M{"_id": objBookID}, updateBook)
	if err != nil {
		// Log the error but proceed, or consider a rollback mechanism
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update book availability"})
	}

	// 6. Respond with success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Book borrowed successfully",
		"book_id": bookID,
	})
}
