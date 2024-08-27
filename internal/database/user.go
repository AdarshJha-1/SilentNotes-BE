package database

import (
	"ama/internal/models"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *service) CheckExistingUser(username, email string) bool {
	filter := bson.M{"$or": []bson.M{
		{"email": email},
		{"username": username},
	}}
	err := UserCollection.FindOne(context.Background(), filter, options.FindOne().SetProjection(bson.M{"password": 0})).Err()

	if err == mongo.ErrNoDocuments {
		return false
	} else if err != nil {
		return false
	}
	return true
}

func (s *service) CreateUser(user models.UserModel) (interface{}, error) {
	result, err := UserCollection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (s *service) GetUser(username string) *models.UserModel {
	var user models.UserModel
	err := UserCollection.FindOne(context.Background(), bson.M{"username": username}, options.FindOne().SetProjection(bson.M{"password": 0})).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		return nil
	}
	return &user
}

func (s *service) VerifyUser(username string) (interface{}, error) {

	filter := bson.M{
		"username": username,
	}
	updateFilter := bson.M{
		"$set": bson.M{"is_verified": true},
		"$unset": bson.M{
			"verify_code":        "",
			"verify_code_expiry": "",
		},
	}

	result, err := UserCollection.UpdateOne(context.Background(), filter, updateFilter)
	if err != nil {
		return nil, err
	}
	fmt.Println(result)
	return result.UpsertedID, nil
}
