// mongo_test.go
package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"explorax-backend/internal/database"
	"explorax-backend/internal/models"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setup(t *testing.T) {
	// Asegúrate de tener MONGO_URI configurado para tests.
	if os.Getenv("MONGO_URI") == "" {
		// Por ejemplo, usa un valor por defecto para test.
		os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	}
	// Conecta a MongoDB.
	database.Connect()
	// Limpia la base de datos de prueba.
	err := database.Client.Database("explorax").Drop(context.Background())
	require.NoError(t, err)
}

func TestConnect(t *testing.T) {
	setup(t)
	require.NotNil(t, database.Client)
	// Opcionalmente, podemos ejecutar un ping a la DB.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := database.Client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
	require.NoError(t, err)
}

func TestInsertAndFindUser(t *testing.T) {
	setup(t)

	user := models.User{
		ID:           primitive.NewObjectID(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}
	err := database.InsertUser(user)
	require.NoError(t, err)

	foundUser, err := database.FindUserByEmail("test@example.com")
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	require.Equal(t, user.Username, foundUser.Username)
}

func TestFindUserByEmailNotFound(t *testing.T) {
	setup(t)

	user, err := database.FindUserByEmail("nonexistent@example.com")
	require.Error(t, err)
	require.Nil(t, user)
}

func TestInsertDuplicateUser(t *testing.T) {
	setup(t)

	user := models.User{
		ID:           primitive.NewObjectID(),
		Username:     "testuser",
		Email:        "duplicate@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
	}

	err := database.InsertUser(user)
	require.NoError(t, err)

	// Try to insert the same user again
	err = database.InsertUser(user)
	require.Error(t, err)
}

func TestInsertMissionAndGetAllMissions(t *testing.T) {
	setup(t)

	mission := models.Mission{
		ID:          primitive.NewObjectID(),
		Title:       "Test Mission",
		Description: "Test Description",
		CreatedAt:   time.Now(),
	}
	err := database.InsertMission(mission)
	require.NoError(t, err)

	missions, err := database.GetAllMissions()
	require.NoError(t, err)
	require.NotEmpty(t, missions)
}

func TestInsertAndUpdateMissionProgress(t *testing.T) {
	setup(t)

	userID := primitive.NewObjectID()
	missionID := primitive.NewObjectID()

	progress := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		MissionID: missionID,
		Status:    "iniciada",
		StartDate: time.Now(),
	}
	err := database.InsertMissionProgress(progress)
	require.NoError(t, err)

	// Actualiza el progreso a "completada".
	err = database.UpdateMissionProgress(userID, missionID)
	require.NoError(t, err)

	// Recupera el progreso y verifica el cambio.
	progs, err := database.GetMissionProgress(userID)
	require.NoError(t, err)
	require.NotEmpty(t, progs)
	require.Equal(t, "completada", progs[0].Status)
}

func TestUpdateMissionProgressInvalidStatus(t *testing.T) {
	setup(t)

	userID := primitive.NewObjectID()
	missionID := primitive.NewObjectID()

	// Try to update progress for non-existent mission
	err := database.UpdateMissionProgress(userID, missionID)
	require.Error(t, err)
}

func TestGetActiveAndCompletedMissions(t *testing.T) {
	setup(t)

	userID := primitive.NewObjectID()

	// Inserta dos progresos: uno iniciado y otro completado.
	progressActive := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		MissionID: primitive.NewObjectID(),
		Status:    "iniciada",
		StartDate: time.Now(),
	}
	progressCompleted := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		MissionID: primitive.NewObjectID(),
		Status:    "completada",
		StartDate: time.Now().Add(-time.Hour),
	}
	err := database.InsertMissionProgress(progressActive)
	require.NoError(t, err)
	err = database.InsertMissionProgress(progressCompleted)
	require.NoError(t, err)

	active, err := database.GetActiveMissions(userID)
	require.NoError(t, err)
	require.Len(t, active, 1)
	require.Equal(t, "iniciada", active[0].Status)

	completed, err := database.GetCompletedMissions(userID)
	require.NoError(t, err)
	require.Len(t, completed, 1)
	require.Equal(t, "completada", completed[0].Status)
}

func TestGetLeaderboard(t *testing.T) {
	setup(t)

	// Para este test, insertamos al menos un usuario y algún progreso.
	// Inserta un usuario.
	user := models.User{
		ID:           primitive.NewObjectID(),
		Username:     "leader",
		Email:        "leader@example.com",
		PasswordHash: "pass",
		CreatedAt:    time.Now(),
	}
	err := database.InsertUser(user)
	require.NoError(t, err)

	// Inserta un progreso completado para ese usuario.
	progress := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		MissionID: primitive.NewObjectID(),
		Status:    "completada",
		StartDate: time.Now().Add(-time.Hour),
	}
	err = database.InsertMissionProgress(progress)
	require.NoError(t, err)

	leaderboard, err := database.GetLeaderboard()
	require.NoError(t, err)
	require.IsType(t, []bson.M{}, leaderboard)
	// Dependiendo de la agregación, el leaderboard podría contener al menos al usuario "leader".
}

func TestGetUserStatistics(t *testing.T) {
	setup(t)

	userID := primitive.NewObjectID()
	// Inserta un progreso completado para este usuario.
	progress := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		MissionID: primitive.NewObjectID(),
		Status:    "completada",
		StartDate: time.Now().Add(-2 * time.Hour),
		EndDate:   time.Now(),
	}
	err := database.InsertMissionProgress(progress)
	require.NoError(t, err)

	stats, err := database.GetUserStatistics(userID)
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Contains(t, stats, "totalCompleted")
	require.Contains(t, stats, "averageDuration")
	require.Contains(t, stats, "progressPercentage")
}

func TestGetMissionByID(t *testing.T) {
	setup(t)

	mission := models.Mission{
		ID:          primitive.NewObjectID(),
		Title:       "Test Mission",
		Description: "Description",
		CreatedAt:   time.Now(),
	}
	err := database.InsertMission(mission)
	require.NoError(t, err)

	fetched, err := database.GetMissionByID(mission.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	require.Equal(t, mission.Title, fetched.Title)
}

func TestGetMissionByIDNotFound(t *testing.T) {
	setup(t)

	nonExistentID := primitive.NewObjectID()
	mission, err := database.GetMissionByID(nonExistentID)
	require.Error(t, err)
	require.Nil(t, mission)
}

func TestGetMissionsOverview(t *testing.T) {
	setup(t)

	// Inserta una misión.
	mission := models.Mission{
		ID:          primitive.NewObjectID(),
		Title:       "Overview Mission",
		Description: "Test overview",
		CreatedAt:   time.Now(),
	}
	err := database.InsertMission(mission)
	require.NoError(t, err)

	// Inserta un progreso completado para la misión.
	userID := primitive.NewObjectID()
	progress := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		MissionID: mission.ID,
		Status:    "completada",
		StartDate: time.Now().Add(-time.Hour),
		EndDate:   time.Now(),
	}
	err = database.InsertMissionProgress(progress)
	require.NoError(t, err)

	overview, err := database.GetMissionsOverview()
	require.NoError(t, err)
	require.NotNil(t, overview)
	// Se puede profundizar en la validación del contenido del overview según la lógica agregada.
}
