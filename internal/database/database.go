package database

import (
	"ama/internal/models"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
	CheckExistingUser(username, email string) bool
	CreateUser(user models.UserModel) (interface{}, error)
	VerifyUser(username string) (interface{}, error)
	GetUser(identifier, projection string) *models.UserModel
	ReVerifyCode(userId primitive.ObjectID, verifyCode int, verifyCodeExpiry time.Time) (interface{}, error)
	ToggleAcceptMessages(isAcceptingMessages bool, userId primitive.ObjectID) bool
	AddMessage(username string, message models.Message) error
	GetMessages(userId primitive.ObjectID) ([]models.Message, error)
	DeleteMessage(userId, messageId primitive.ObjectID) error
}

type service struct {
	db *mongo.Client
}

var (
	UserCollection    *mongo.Collection
	MessageCollection *mongo.Collection
)

var (
	dbURI       = os.Getenv("MONGO_DB_URI")
	database    = os.Getenv("DB_NAME")
	userColl    = os.Getenv("USER_COLL")
	messageColl = os.Getenv("MESSAGE_COLL")
)

func New() Service {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dbURI))

	if err != nil {
		log.Fatal(err)

	}
	UserCollection = client.Database(database).Collection(userColl)
	MessageCollection = client.Database(database).Collection(messageColl)

	return &service{
		db: client,
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		log.Fatal(fmt.Printf("db down: %v", err))
	}

	return map[string]string{
		"message": "It's healthy",
	}
}
