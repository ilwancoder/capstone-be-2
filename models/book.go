package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	Author    string             `bson:"author"`
	Synopsis  string             `bson:"synopsis"`
	Quantity  int                `bson:"quantity"`
	Available int                `bson:"available"`
	Genre     string             `bson:"genre"`
	AgeGroup  string             `bson:"age_group"`
	CoverUrl  string             `bson:"cover_url"`
}
