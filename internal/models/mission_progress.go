// /internal/models/mission_progress.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MissionProgress almacena el estado de una misión para un usuario específico.
type MissionProgress struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	MissionID primitive.ObjectID `bson:"missionId" json:"missionId"`
	Status    string             `bson:"status" json:"status"` // "iniciada" o "completada"
	StartDate time.Time          `bson:"startDate" json:"startDate"`
	EndDate   time.Time          `bson:"endDate,omitempty" json:"endDate,omitempty"`
}
