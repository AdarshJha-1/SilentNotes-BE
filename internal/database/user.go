package database

import (
	"ama/internal/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (s *service) GetUser(identifier, projection string) *models.UserModel {
	var user models.UserModel
	filter := bson.M{"$or": []bson.M{
		{"email": identifier},
		{"username": identifier},
	}}
	var err error
	if projection == "" {
		err = UserCollection.FindOne(context.Background(), filter).Decode(&user)
	} else {
		err = UserCollection.FindOne(context.Background(), filter, options.FindOne().SetProjection(bson.M{projection: 0})).Decode(&user)
	}
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
	return result.UpsertedID, nil
}

func (s *service) ReVerifyCode(userId primitive.ObjectID, verifyCode int, verifyCodeExpiry time.Time) (interface{}, error) {
	updateFilter := bson.M{
		"$set": bson.M{
			"verify_code":        verifyCode,
			"verify_code_expiry": verifyCodeExpiry,
		},
	}
	result, err := UserCollection.UpdateByID(context.Background(), userId, updateFilter)
	if err != nil {
		return nil, err
	}
	return result.UpsertedID, nil
}

func (s *service) ToggleAcceptMessages(isAcceptingMessages bool, userId primitive.ObjectID) bool {

	updateFilter := bson.M{
		"$set": bson.M{
			"is_accepting_messages": isAcceptingMessages,
		},
	}
	result, err := UserCollection.UpdateByID(context.Background(), userId, updateFilter)
	if err != nil {
		return false
	}

	if result.MatchedCount == 0 {
		return false
	}
	return true
}

func (s *service) AddMessage(username string, message models.Message) error {
	filter := bson.M{"username": username}
	updateFilter := bson.M{
		"$push": bson.M{
			"messages": message,
		},
	}
	result, err := UserCollection.UpdateOne(context.Background(), filter, updateFilter)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (s *service) GetMessages(userId primitive.ObjectID) ([]models.Message, error) {

	var user models.UserModel
	var userMessages []models.Message
	result := UserCollection.FindOne(context.Background(), bson.M{"_id": userId}, options.FindOne().SetProjection(bson.M{
		"_id":                   0,
		"username":              0,
		"email":                 0,
		"password":              0,
		"is_accepting_messages": 0,
		"is_verified":           0,
	}))
	err := result.Decode(&user)
	userMessages = user.Messages
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(userMessages) == 0 {
		return nil, nil
	}
	return userMessages, nil
}

func (s *service) DeleteMessage(userId, messageId primitive.ObjectID) error {
	updateFilter := bson.M{
		"$pull": bson.M{
			"messages": bson.M{
				"_id": messageId,
			},
		},
	}
	result, err := UserCollection.UpdateByID(context.Background(), userId, updateFilter)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}
