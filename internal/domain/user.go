package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email" binding:"required"`
	Photo     string             `json:"photo" bson:"photo"`
	Name      string             `json:"name" bson:"name" binding:"required"`
	CreatedAt int64              `json:"created_at" bson:"created_at" binding:"required"`
}

type UserUpdate struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Photo string             `json:"photo" bson:"photo"`
	Name  string             `json:"name" bson:"name" binding:"required"`
}
