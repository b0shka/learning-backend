package repository

import (
	"context"
	"errors"

	"github.com/b0shka/backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UsersRepo struct {
	db *mongo.Collection
}

func NewUsersRepo(db *mongo.Database) *UsersRepo {
	return &UsersRepo{
		db: db.Collection(usersCollection),
	}
}

func (r *UsersRepo) AddVerifyEmail(ctx context.Context, verifyEmail domain.VerifyEmail) error {
	_, err := r.db.Database().Collection(verifyEmailsCollection).InsertOne(ctx, verifyEmail)
	return err
}

func (r *UsersRepo) GetVerifyEmail(ctx context.Context, email, code string) (domain.VerifyEmail, error) {
	var verifyEmail domain.VerifyEmail
	filter := bson.M{
		"email":       email,
		"secret_code": code,
	}

	if err := r.db.Database().Collection(verifyEmailsCollection).FindOne(ctx, filter).Decode(&verifyEmail); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.VerifyEmail{}, domain.ErrSecretCodeInvalid
		}
		return domain.VerifyEmail{}, err
	}

	return verifyEmail, nil
}

func (r *UsersRepo) RemoveVerifyEmail(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.db.Database().Collection(verifyEmailsCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *UsersRepo) Create(ctx context.Context, user domain.User) error {
	_, err := r.db.InsertOne(ctx, user)
	return err
}

func (r *UsersRepo) Get(ctx context.Context, identifier interface{}) (domain.User, error) {
	var user domain.User
	var filter bson.M

	switch identifier.(type) {
	case string:
		filter = bson.M{
			"email": identifier,
		}
	case primitive.ObjectID:
		filter = bson.M{
			"_id": identifier,
		}
	default:
		return domain.User{}, domain.ErrIdentifier
	}

	if err := r.db.FindOne(ctx, filter).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *UsersRepo) Update(ctx context.Context, id primitive.ObjectID, user domain.UserUpdate) error {
	_, err := r.db.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": user})
	return err
}
