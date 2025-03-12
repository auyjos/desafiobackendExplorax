// /internal/database/mongo.go
package database

import (
	"context"
	"log"
	"time"

	"explorax-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// Connect establece la conexión a MongoDB usando ServerAPIOptions.
func Connect() {

	mongoURI := "mongodb+srv://jose:1234@etm.nwysozv.mongodb.net/?retryWrites=true&w=majority&appName=ETM"
	if mongoURI == "" {
		log.Fatal("MONGO_URI no está configurada. Revisa tu .env o variables de entorno.")
	}

	// Configura el ServerAPIOptions con la versión estable de la API (ServerAPIVersion1)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal("Error conectando a MongoDB: ", err)
	}

	// Realiza un ping para confirmar la conexión
	if err := client.Database("admin").RunCommand(ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		log.Fatal("No se pudo hacer ping a MongoDB: ", err)
	}

	Client = client
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")
}

func GetUserCollection() *mongo.Collection {
	return Client.Database("explorax").Collection("users")
}

func GetMissionProgressCollection() *mongo.Collection {
	return Client.Database("explorax").Collection("mission_progress")
}

// InsertUser inserta un nuevo usuario en la base de datos.
func InsertUser(user models.User) error {
	collection := GetUserCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, user)
	return err
}

// FindUserByEmail busca un usuario por email.
func FindUserByEmail(email string) (*models.User, error) {
	collection := GetUserCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// InsertMission inserta una misión.
func InsertMission(mission models.Mission) error {
	collection := Client.Database("explorax").Collection("missions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, mission)
	return err
}

// GetAllMissions obtiene todas las misiones.
func GetAllMissions() ([]models.Mission, error) {
	collection := Client.Database("explorax").Collection("missions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var missions []models.Mission
	if err = cursor.All(ctx, &missions); err != nil {
		return nil, err
	}
	return missions, nil
}

// InsertMissionProgress inserta un nuevo documento de progreso de misión.
func InsertMissionProgress(progress models.MissionProgress) error {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, progress)
	return err
}

// UpdateMissionProgress actualiza el progreso de una misión a "completada" y registra la fecha final.
func UpdateMissionProgress(userID, missionID string) error {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"userId": userID, "missionId": missionID, "status": "iniciada"}
	update := bson.M{
		"$set": bson.M{
			"status":  "completada",
			"endDate": time.Now(),
		},
	}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// GetMissionProgress obtiene todos los documentos de progreso de misión para un usuario.
func GetMissionProgress(userID string) ([]models.MissionProgress, error) {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	var progress []models.MissionProgress
	if err = cursor.All(ctx, &progress); err != nil {
		return nil, err
	}
	return progress, nil
}

// GetActiveMissions retorna las misiones con estado "iniciada" para un usuario.
func GetActiveMissions(userID string) ([]models.MissionProgress, error) {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{"userId": userID, "status": "iniciada"})
	if err != nil {
		return nil, err
	}
	var active []models.MissionProgress
	if err = cursor.All(ctx, &active); err != nil {
		return nil, err
	}
	return active, nil
}

// GetCompletedMissions retorna las misiones con estado "completada" para un usuario.
func GetCompletedMissions(userID string) ([]models.MissionProgress, error) {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{"userId": userID, "status": "completada"})
	if err != nil {
		return nil, err
	}
	var completed []models.MissionProgress
	if err = cursor.All(ctx, &completed); err != nil {
		return nil, err
	}
	return completed, nil
}

// GetLeaderboard retorna un ranking de usuarios basado en el número de misiones completadas.
func GetLeaderboard() ([]bson.M, error) {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"status": "completada"}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":            "$userId",
			"completedCount": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"completedCount": -1}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var leaderboard []bson.M
	if err = cursor.All(ctx, &leaderboard); err != nil {
		return nil, err
	}
	return leaderboard, nil
}

// GetUserStatistics retorna estadísticas para un usuario, como total de misiones completadas y duración promedio.
func GetUserStatistics(userID string) (bson.M, error) {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Total de misiones completadas
	total, err := collection.CountDocuments(ctx, bson.M{"userId": userID, "status": "completada"})
	if err != nil {
		return nil, err
	}

	// Cálculo de duración promedio de misiones completadas
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"userId": userID, "status": "completada"}}},
		bson.D{{Key: "$project", Value: bson.M{
			"duration": bson.M{
				"$subtract": []interface{}{"$endDate", "$startDate"},
			},
		}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":             nil,
			"averageDuration": bson.M{"$avg": "$duration"},
		}}},
	}
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	var avgDuration interface{}
	if len(results) > 0 {
		avgDuration = results[0]["averageDuration"]
	} else {
		avgDuration = 0
	}

	return bson.M{
		"totalCompleted":  total,
		"averageDuration": avgDuration,
	}, nil
}
