package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Repositories struct {
}

func NewRepositories(db *mongo.Database) *Repositories {
	return &Repositories{}
}
