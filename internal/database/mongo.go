// /internal/database/mongo.go
package database

import (
	"context"
	"log"
	"os"
	"time"

	"explorax-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// Connect establece la conexión a MongoDB usando ServerAPIOptions.
func Connect() {

	mongoURI := os.Getenv("MONGO_URI")
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
func UpdateMissionProgress(userID, missionID primitive.ObjectID) error {
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
func GetMissionProgress(userID primitive.ObjectID) ([]models.MissionProgress, error) {
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
func GetActiveMissions(userID primitive.ObjectID) ([]models.MissionProgress, error) {
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
func GetCompletedMissions(userID primitive.ObjectID) ([]models.MissionProgress, error) {
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

// GetLeaderboard retorna un ranking de todos los usuarios basado en misiones completadas.
// Incluye a los usuarios con 0 completadas.
func GetLeaderboard() ([]bson.M, error) {
	userCollection := Client.Database("explorax").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		// 1) Realiza un lookup para unir con "mission_progress"
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "mission_progress",
			"localField":   "_id",    // _id de users
			"foreignField": "userId", // userId en mission_progress
			"as":           "progress",
		}}},
		// 2) Desenrolla el array "progress" para contar cada registro individualmente
		bson.D{{Key: "$unwind", Value: bson.M{
			"path":                       "$progress",
			"preserveNullAndEmptyArrays": true,
		}}},
		// 3) Agrupa por usuario y suma 1 para cada misión completada
		bson.D{{Key: "$group", Value: bson.M{
			"_id":      "$_id",
			"username": bson.M{"$first": "$username"},
			"email":    bson.M{"$first": "$email"},
			"completedCount": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$progress.status", "completada"}},
						1,
						0,
					},
				},
			},
		}}},
		// 4) Ordena de mayor a menor por completedCount
		bson.D{{Key: "$sort", Value: bson.M{"completedCount": -1}}},
	}

	cursor, err := userCollection.Aggregate(ctx, pipeline)
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
func GetUserStatistics(userID primitive.ObjectID) (bson.M, error) {
	// Contexto para las consultas.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Total de misiones completadas por el usuario.
	progressCollection := GetMissionProgressCollection()
	totalCompleted, err := progressCollection.CountDocuments(ctx, bson.M{
		"userId": userID,
		"status": "completada",
	})
	if err != nil {
		return nil, err
	}

	// 2. Calcular duración promedio de misiones completadas.
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{
			"userId": userID,
			"status": "completada",
		}}},
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
	cursor, err := progressCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	var avgDuration any
	if len(results) > 0 {
		avgDuration = results[0]["averageDuration"]
	} else {
		avgDuration = 0
	}

	// 3. Total de misiones disponibles en el sistema.
	missionsCollection := Client.Database("explorax").Collection("missions")
	totalMissions, err := missionsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// 4. Calcular porcentaje de avance.
	progressPercentage := 0.0
	if totalMissions > 0 {
		progressPercentage = (float64(totalCompleted) / float64(totalMissions)) * 100
	}

	// Retorna estadísticas del usuario.
	return bson.M{
		"totalCompleted":     totalCompleted,
		"averageDuration":    avgDuration,
		"progressPercentage": progressPercentage,
	}, nil
}

func GetMissionByID(id primitive.ObjectID) (*models.Mission, error) {
	collection := Client.Database("explorax").Collection("missions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var mission models.Mission
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&mission)
	if err != nil {
		return nil, err
	}
	return &mission, nil
}

// GetMissionsOverview calcula estadísticas globales:
// - Misión más popular (mayor número de completadas).
// - Tiempo promedio de finalización por misión.
func GetMissionsOverview() (bson.M, error) {
	collection := GetMissionProgressCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pipeline para la misión más popular:
	pipelineMostPopular := mongo.Pipeline{
		// Filtra solo las misiones completadas.
		bson.D{{Key: "$match", Value: bson.M{"status": "completada"}}},
		// Agrupa por missionId y cuenta cuántas veces se completó.
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$missionId",
			"count": bson.M{"$sum": 1},
		}}},
		// Ordena de mayor a menor.
		bson.D{{Key: "$sort", Value: bson.M{"count": -1}}},
		// Toma solo la misión con más completadas.
		bson.D{{Key: "$limit", Value: 1}},
		// Lookup para traer la misión de la colección "missions".
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "missions",
			"localField":   "_id", // ya es ObjectID
			"foreignField": "_id",
			"as":           "mission",
		}}},
		// Desenrolla el array "mission".
		bson.D{{Key: "$unwind", Value: "$mission"}},
		// Proyecta el resultado final.
		bson.D{{Key: "$project", Value: bson.M{
			"_id":       0,
			"missionId": "$_id",
			"count":     1,
			"mission":   1,
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipelineMostPopular)
	if err != nil {
		return nil, err
	}
	var popularResult []bson.M
	if err = cursor.All(ctx, &popularResult); err != nil {
		return nil, err
	}

	var mostPopular interface{}
	if len(popularResult) > 0 {
		mostPopular = popularResult[0]
	} else {
		mostPopular = nil
	}

	// Pipeline para calcular el promedio de duración por misión:
	pipelineAvgTime := mongo.Pipeline{
		// Filtra solo las misiones completadas.
		bson.D{{Key: "$match", Value: bson.M{"status": "completada"}}},
		// Agrupa por missionId y calcula el promedio de (endDate - startDate).
		bson.D{{Key: "$group", Value: bson.M{
			"_id": "$missionId",
			"averageDuration": bson.M{
				"$avg": bson.M{"$subtract": []interface{}{"$endDate", "$startDate"}},
			},
			"count": bson.M{"$sum": 1},
		}}},
		// Lookup para traer la información de la misión.
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "missions",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "mission",
		}}},
		bson.D{{Key: "$unwind", Value: "$mission"}},

		bson.D{{Key: "$project", Value: bson.M{
			"_id":             0,
			"missionId":       "$_id",
			"averageDuration": 1,
			"count":           1,
			"mission":         1,
		}}},
	}

	cursor2, err := collection.Aggregate(ctx, pipelineAvgTime)
	if err != nil {
		return nil, err
	}
	var avgResults []bson.M
	if err = cursor2.All(ctx, &avgResults); err != nil {
		return nil, err
	}

	// Estructura final del overview
	overview := bson.M{
		"mostPopularMission": mostPopular,
		"avgCompletionTimes": avgResults,
	}

	return overview, nil
}
