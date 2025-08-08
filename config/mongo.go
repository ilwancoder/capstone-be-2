package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var DB *mongo.Client

// ConnectDB initializes the MongoDB connection
func ConnectDB() {
	// Use the MONGODB_URI from the environment variables
	mongoDbURI := os.Getenv("MONGODB_URI")
	if mongoDbURI == "" {
		log.Fatal("MONGODB_URI is not set in the .env file")
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoDbURI)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("Error creating MongoDB client: %v", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	// Ping the database to verify the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
	}

	fmt.Println("🎉 Connected to MongoDB successfully!")
	DB = client
}

// GetCollection returns a collection from the database
func GetCollection(collectionName string) *mongo.Collection {
	databaseName := os.Getenv("MONGODB_DBNAME")
	if databaseName == "" {
		databaseName = "library"
	}
	return DB.Database(databaseName).Collection(collectionName)
}
