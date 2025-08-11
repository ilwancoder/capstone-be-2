package controllers

import (
	"context"
	"strings"
	"time"

	"capstone-be-2/config"
	"capstone-be-2/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAllBooks(c *fiber.Ctx) error {
	booksColl := config.GetCollection("books")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := booksColl.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Error fetching books"})
	}
	defer cur.Close(ctx)

	var books []models.Book
	if err := cur.All(ctx, &books); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Error decoding books"})
	}
	return c.JSON(books)
}

func GetBookByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid book ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	books := config.GetCollection("books")

	var book models.Book
	if err := books.FindOne(ctx, bson.M{"_id": objID}).Decode(&book); err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Book not found"})
	}
	return c.JSON(book)
}

// --- Create (opsional, untuk seed/admin) ---
func CreateBook(c *fiber.Ctx) error {
	var body struct {
		Title    string `json:"title"`
		Author   string `json:"author"`
		Synopsis string `json:"synopsis"`
		Quantity int    `json:"quantity"`
		CoverUrl string `json:"cover_url"`
		Genre    string `json:"genre"`
		AgeGroup string `json:"age_group"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "cannot parse JSON"})
	}
	if strings.TrimSpace(body.Title) == "" || strings.TrimSpace(body.Author) == "" {
		return c.Status(400).JSON(fiber.Map{"message": "title & author are required"})
	}
	if body.Quantity <= 0 {
		body.Quantity = 1
	}

	book := models.Book{
		ID:        primitive.NewObjectID(),
		Title:     body.Title,
		Author:    body.Author,
		Synopsis:  body.Synopsis,
		Quantity:  body.Quantity,
		Available: body.Quantity,
		CoverUrl:  body.CoverUrl,
		Genre:     body.Genre,
		AgeGroup:  body.AgeGroup,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := config.GetCollection("books").InsertOne(ctx, book); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "could not create book"})
	}
	return c.Status(201).JSON(book)
}

// --- Borrow (protected) ---
func BorrowBook(c *fiber.Ctx) error {
	tok := c.Locals("user").(*jwt.Token)
	claims := tok.Claims.(jwt.MapClaims)
	idHex, ok := claims["id"].(string)
	if !ok || idHex == "" {
		return c.Status(401).JSON(fiber.Map{"message": "Invalid token"})
	}
	userID, _ := primitive.ObjectIDFromHex(idHex)

	bookIDHex := c.Params("id")
	bookID, err := primitive.ObjectIDFromHex(bookIDHex)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid book ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	users := config.GetCollection("users")
	books := config.GetCollection("books")

	// cek buku
	var book models.Book
	if err := books.FindOne(ctx, bson.M{"_id": bookID}).Decode(&book); err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Book not found"})
	}
	if book.Available <= 0 {
		return c.Status(409).JSON(fiber.Map{"message": "Book is not available"})
	}

	// cek user & duplikat
	var user models.User
	if err := users.FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "User not found"})
	}
	for _, b := range user.BorrowedBooks {
		if b == bookID {
			return c.Status(400).JSON(fiber.Map{"message": "Book already borrowed"})
		}
	}

	// update user & book
	if _, err := users.UpdateByID(ctx, userID, bson.M{"$push": bson.M{"borrowed_books": bookID}}); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to update user"})
	}
	if _, err := books.UpdateByID(ctx, bookID, bson.M{"$inc": bson.M{"available": -1}}); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to update book availability"})
	}
	return c.JSON(fiber.Map{"message": "Book borrowed", "book_id": bookIDHex})
}

// --- Return (protected) ---
func ReturnBook(c *fiber.Ctx) error {
	tok := c.Locals("user").(*jwt.Token)
	claims := tok.Claims.(jwt.MapClaims)
	idHex, ok := claims["id"].(string)
	if !ok || idHex == "" {
		return c.Status(401).JSON(fiber.Map{"message": "Invalid token"})
	}
	userID, _ := primitive.ObjectIDFromHex(idHex)

	bookIDHex := c.Params("id")
	bookID, err := primitive.ObjectIDFromHex(bookIDHex)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid book ID"})
	}

	var body struct {
		Summary string `json:"summary"`
		Comment string `json:"comment"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "cannot parse JSON"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	users := config.GetCollection("users")
	books := config.GetCollection("books")
	summaries := config.GetCollection("summaries")

	// pastikan user pernah meminjam
	if err := users.FindOne(ctx, bson.M{"_id": userID, "borrowed_books": bookID}).Err(); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Book not in user's borrowed list"})
	}

	// keluarkan dari list pinjaman
	if _, err := users.UpdateByID(ctx, userID, bson.M{"$pull": bson.M{"borrowed_books": bookID}}); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to update user borrowed list"})
	}
	// kembalikan stok
	if _, err := books.UpdateByID(ctx, bookID, bson.M{"$inc": bson.M{"available": 1}}); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to restore availability"})
	}

	// hitung poin sederhana: 1 poin per paragraf (dipisah newline ganda)
	points := 0
	trim := strings.TrimSpace(body.Summary)
	if trim != "" {
		paras := strings.Split(trim, "\n\n")
		points = len(paras)
		// simpan ringkasan (opsional)
		_, _ = summaries.InsertOne(ctx, bson.M{
			"_id":            primitive.NewObjectID(),
			"user_id":        userID,
			"book_id":        bookID,
			"summary_text":   body.Summary,
			"comment_text":   body.Comment,
			"points_awarded": points,
			"created_at":     time.Now(),
		})
	}
	if points > 0 {
		if _, err := users.UpdateByID(ctx, userID, bson.M{"$inc": bson.M{"points": points}}); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Failed to update points"})
		}
	}

	return c.JSON(fiber.Map{"message": "Book returned", "points_awarded": points})
}

func GetMyBorrowed(c *fiber.Ctx) error {
	tok := c.Locals("user").(*jwt.Token)
	claims := tok.Claims.(jwt.MapClaims)
	idHex, _ := claims["id"].(string)
	userID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "invalid user id"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// ambil list ObjectID buku dari user
	var user models.User
	if err := config.GetCollection("users").
		FindOne(ctx, bson.M{"_id": userID}).
		Decode(&user); err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "user not found"})
	}

	if len(user.BorrowedBooks) == 0 {
		return c.JSON([]models.Book{}) // kosong
	}

	// ambil detail buku
	cur, err := config.GetCollection("books").
		Find(ctx, bson.M{"_id": bson.M{"$in": user.BorrowedBooks}})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "db error"})
	}
	defer cur.Close(ctx)

	var books []models.Book
	if err := cur.All(ctx, &books); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "decode error"})
	}
	return c.JSON(books)
}

// GET /api/leaderboard  (public)
func GetLeaderboard(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "points", Value: -1}}).SetLimit(10)
	cur, err := config.GetCollection("users").Find(ctx, bson.M{}, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "db error"})
	}
	defer cur.Close(ctx)

	type Entry struct {
		Name   string `json:"name"`
		Points int    `json:"points"`
	}
	var out []Entry

	for cur.Next(ctx) {
		var u models.User
		if err := cur.Decode(&u); err != nil {
			continue
		}
		out = append(out, Entry{Name: u.Name, Points: u.Points})
	}
	return c.JSON(out)
}
