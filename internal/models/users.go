package models

import "time"

type User struct {
	ID           string    `json:"id,omitempty" bson:"_id,omitempty"`
	Username     string    `json:"username" bson:"username"`
	Email        string    `json:"email" bson:"email"`
	PasswordHash string    `json:"-" bson:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
}
