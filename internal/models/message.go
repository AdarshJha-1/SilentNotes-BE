package models

import "time"

type Message struct {
	Content   string    `json:"content" bson:"content"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at"`
}
