// /internal/models/mission_progress.go
package models

import "time"

// MissionProgress almacena el estado de una misión para un usuario específico.
type MissionProgress struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    string    `bson:"userId" json:"userId"`
	MissionID string    `bson:"missionId" json:"missionId"`
	Status    string    `bson:"status" json:"status"` // "iniciada" o "completada"
	StartDate time.Time `bson:"startDate" json:"startDate"`
	EndDate   time.Time `bson:"endDate,omitempty" json:"endDate,omitempty"`
}
