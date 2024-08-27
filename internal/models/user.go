package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserModel struct {
	ID                  primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Username            string             `json:"username" bson:"username" validate:"required,min=3,max=30"`
	Email               string             `json:"email" bson:"email" validate:"required,email"`
	Password            string             `json:"password" bson:"password" validate:"required"`
	IsVerified          bool               `json:"is_verified,omitempty" bson:"is_verified,omitempty"`
	IsAcceptingMessages bool               `json:"is_accepting_messages,omitempty" bson:"is_accepting_messages,omitempty"`
	VerifyCode          int                `json:"verify_code,omitempty" bson:"verify_code,omitempty"`
	VerifyCodeExpiry    time.Time          `json:"verify_code_expiry,omitempty" bson:"verify_code_expiry,omitempty"`
	Messages            []Message          `json:"messages,omitempty" bson:"messages,omitempty"`
}

type SingInModel struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}
