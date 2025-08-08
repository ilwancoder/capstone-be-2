package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BorrowRecord struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	UserID        primitive.ObjectID `bson:"user_id"`
	BookID        primitive.ObjectID `bson:"book_id"`
	BorrowDate    time.Time          `bson:"borrow_date"`
	ReturnDate    *time.Time         `bson:"return_date,omitempty"` // Pointer to allow for nil value
	SummaryText   string             `bson:"summary_text,omitempty"`
	PointsAwarded int                `bson:"points_awarded,omitempty"`
}
