package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type VerifyEmailConfig struct {
	Subject string
	Content string
}

type VerifyEmail struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Email          string             `binding:"required"`
	SecretCodeHash string             `bson:"secret_code" binding:"required"`
	ExpiredAt      int64              `bson:"expired_at" binding:"required"`
}
