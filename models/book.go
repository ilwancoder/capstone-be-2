package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// models/book.go
type Book struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Title     string             `bson:"title" json:"title"`
	Author    string             `bson:"author" json:"author"`
	Synopsis  string             `bson:"synopsis" json:"synopsis"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Available int                `bson:"available" json:"available"`
	CoverUrl  string             `bson:"cover_url" json:"cover_url,omitempty"`
	Genre     string             `bson:"genre" json:"genre,omitempty"`
	AgeGroup  string             `bson:"age_group" json:"age_group,omitempty"`
	ISBN      string             `bson:"isbn" json:"isbn,omitempty"`
}
