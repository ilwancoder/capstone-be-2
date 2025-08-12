// cmd/seed/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"capstone-be-2/config"
	"capstone-be-2/models"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	// cari .env di root project
	_ = godotenv.Load(".env")
}
func main() {
	config.ConnectDB()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	books := config.GetCollection("books")

	// (opsional) index untuk search cepat
	createIndexes(ctx, books)

	// dataset – available di-set sama dengan quantity
	dataset := []models.Book{
		{ID: primitive.NewObjectID(), Title: "Clean Code", Author: "Robert C. Martin", Synopsis: "A handbook of agile software craftsmanship.", Quantity: 5, Available: 5, Genre: "Programming", AgeGroup: "Adult", ISBN: "9780132350884"},
		{ID: primitive.NewObjectID(), Title: "The Pragmatic Programmer", Author: "Andrew Hunt, David Thomas", Synopsis: "Pragmatic practices for software developers.", Quantity: 4, Available: 4, Genre: "Programming", AgeGroup: "Adult", ISBN: "9780201616224"},
		{ID: primitive.NewObjectID(), Title: "Design Patterns", Author: "Erich Gamma, Richard Helm, Ralph Johnson, John Vlissides", Synopsis: "Elements of Reusable Object-Oriented Software.", Quantity: 3, Available: 3, Genre: "Software Engineering", AgeGroup: "Adult", ISBN: "9780201633610"},
		{ID: primitive.NewObjectID(), Title: "Refactoring", Author: "Martin Fowler", Synopsis: "Improving the design of existing code.", Quantity: 4, Available: 4, Genre: "Programming", AgeGroup: "Adult", ISBN: "9780201485677"},
		{ID: primitive.NewObjectID(), Title: "You Don't Know JS Yet", Author: "Kyle Simpson", Synopsis: "Deep dive into modern JavaScript.", Quantity: 6, Available: 6, Genre: "Programming", AgeGroup: "Adult", ISBN: "9781098124045"},
		{ID: primitive.NewObjectID(), Title: "The Hobbit", Author: "J.R.R. Tolkien", Synopsis: "Bilbo's unexpected journey with the dwarves.", Quantity: 4, Available: 4, Genre: "Fantasy", AgeGroup: "Teen", ISBN: "9780007458424"},
		{ID: primitive.NewObjectID(), Title: "Harry Potter and the Sorcerer's Stone", Author: "J.K. Rowling", Synopsis: "Harry discovers the wizarding world.", Quantity: 5, Available: 5, Genre: "Fantasy", AgeGroup: "Teen", ISBN: "9780439708180"},
		{ID: primitive.NewObjectID(), Title: "The Lord of the Rings", Author: "J.R.R. Tolkien", Synopsis: "Epic quest to destroy the One Ring.", Quantity: 2, Available: 2, Genre: "Fantasy", AgeGroup: "Adult", ISBN: "9780618640157"},
		{ID: primitive.NewObjectID(), Title: "Atomic Habits", Author: "James Clear", Synopsis: "Tiny changes, remarkable results.", Quantity: 5, Available: 5, Genre: "Self-Help", AgeGroup: "Adult", ISBN: "9780735211292"},
		{ID: primitive.NewObjectID(), Title: "Deep Work", Author: "Cal Newport", Synopsis: "Rules for focused success in a distracted world.", Quantity: 3, Available: 3, Genre: "Productivity", AgeGroup: "Adult", ISBN: "9781455586691"},
	}

	// Upsert per (title+author) agar tidak dobel kalau script dijalankan berkali-kali
	upserts := 0
	for _, b := range dataset {
		if b.Available == 0 { // jaga-jaga
			b.Available = b.Quantity
		}
		filter := bson.M{"title": b.Title, "author": b.Author}
		update := bson.M{"$setOnInsert": b}
		opts := options.Update().SetUpsert(true)

		res, err := books.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("upsert failed for %q: %v\n", b.Title, err)
			continue
		}
		if res.UpsertedCount > 0 {
			upserts++
			log.Printf("seeded: %s (%s)\n", b.Title, b.Author)
		} else {
			log.Printf("exists: %s (%s) – skipped\n", b.Title, b.Author)
		}
	}

	fmt.Printf("\nDone. Inserted/Upserted: %d item(s).\n", upserts)
}

func createIndexes(ctx context.Context, coll *mongo.Collection) {
	// index text untuk title+author (membantu search)
	idx := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "author", Value: "text"},
		},
		Options: options.Index().SetName("text_title_author"),
	}
	_, _ = coll.Indexes().CreateOne(ctx, idx)
}
