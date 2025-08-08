package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty"`
	Name          string               `bson:"name"`
	Email         string               `bson:"email"`
	PasswordHash  string               `bson:"password_hash"`
	Points        int                  `bson:"points"`
	BorrowedBooks []primitive.ObjectID `bson:"borrowed_books"`
}
