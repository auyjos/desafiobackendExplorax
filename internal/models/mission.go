// /internal/models/mission.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mission representa la estructura de una misión en la base de datos.
type Mission struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}
