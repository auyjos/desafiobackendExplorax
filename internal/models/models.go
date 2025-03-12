// /internal/models/mission.go
package models

import "time"

// Mission representa la estructura de una misión en la base de datos.
type Mission struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string    `bson:"title" json:"title"`
	Description string    `bson:"description" json:"description"`
	CreatedAt   time.Time `bson:"createdAt" json:"createdAt"`
}
