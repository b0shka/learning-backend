package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type VerifyEmailConfig struct {
	Subject string
	Content string
}

type VerifyEmail struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email      string             `json:"email" binding:"required"`
	SecretCode string             `json:"secret_code" bson:"secret_code" binding:"required"`
	ExpiredAt  int64              `json:"expired_at" bson:"expired_at" binding:"required"`
}
