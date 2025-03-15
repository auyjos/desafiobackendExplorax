// /internal/handlers/missions.go
package handlers

import (
	"net/http"
	"time"

	"explorax-backend/internal/database"
	"explorax-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// StartMission - POST /missions/start
// Inicia una misión registrando el user_id (del JWT), mission_id, fecha de inicio y estado "iniciada".
func StartMission(c *gin.Context) {
	// Obtener el user_id del contexto y convertirlo a string
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de ID de usuario incorrecto"})
		return
	}

	// Convertir el userID de string a ObjectID
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al convertir el ID del usuario"})
		return
	}

	// El input ahora vincula directamente un primitive.ObjectID para MissionID
	var input struct {
		MissionID primitive.ObjectID `json:"mission_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	progress := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    userObjID,
		MissionID: input.MissionID,
		Status:    "iniciada",
		StartDate: time.Now(),
	}

	if err := database.InsertMissionProgress(progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al iniciar la misión"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Misión iniciada"})
}

// CompleteMission - POST /missions/complete
// Actualiza el progreso de una misión a "completada" y registra la fecha de finalización.
func CompleteMission(c *gin.Context) {
	// Obtener el user_id del contexto y convertirlo a ObjectID
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	userIDStr, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de ID de usuario incorrecto"})
		return
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al convertir el ID del usuario"})
		return
	}

	// Vincular y convertir el MissionID del body a ObjectID
	var input struct {
		MissionID string `json:"mission_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}
	missionObjID, err := primitive.ObjectIDFromHex(input.MissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de misión inválido"})
		return
	}

	// Actualizar el progreso; la función UpdateMissionProgress usa un filtro
	// que solo coincide si el status es "iniciada"
	err = database.UpdateMissionProgress(userObjID, missionObjID)
	if err != nil {
		// Si no se encontró ningún documento, se asume que la misión no fue iniciada
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No has iniciado esta misión, no puedes completarla"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al completar la misión"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Misión completada"})
}

// GetProgress - GET /missions/progress
// Devuelve el progreso de misiones (tanto iniciadas como completadas) del usuario.
func GetProgress(c *gin.Context) {
	// Obtener el user_id del contexto y convertirlo a ObjectID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de ID de usuario incorrecto"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al convertir el ID del usuario"})
		return
	}

	progress, err := database.GetMissionProgress(userObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el progreso"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

func GetActiveMissions(c *gin.Context) {
	// Obtener el user_id del contexto y convertirlo a ObjectID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de ID de usuario incorrecto"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al convertir el ID del usuario"})
		return
	}

	active, err := database.GetActiveMissions(userObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener misiones activas"})
		return
	}

	c.JSON(http.StatusOK, active)
}

func GetCompletedMissions(c *gin.Context) {
	// Obtener el user_id del contexto y convertirlo a ObjectID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de ID de usuario incorrecto"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al convertir el ID del usuario"})
		return
	}

	completed, err := database.GetCompletedMissions(userObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener misiones completadas"})
		return
	}

	c.JSON(http.StatusOK, completed)
}

// GetLeaderboard - GET /missions/leaderboard
// Devuelve un ranking de usuarios basado en el número de misiones completadas.
func GetLeaderboard(c *gin.Context) {
	leaderboard, err := database.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el leaderboard"})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

// GetStatistics - GET /missions/statistics
// Devuelve estadísticas generales del usuario, como total de misiones completadas y tiempo promedio de finalización.
func GetStatistics(c *gin.Context) {
	// Extraer el user_id del contexto
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// Convertir el user_id (que se espera como string) a ObjectID
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de ID de usuario incorrecto"})
		return
	}
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al convertir el ID del usuario"})
		return
	}

	// Llamar a GetUserStatistics con el ObjectID
	stats, err := database.GetUserStatistics(userObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener estadísticas"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateMission - POST /missions (Crea una nueva misión)
func CreateMission(c *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	mission := models.Mission{
		ID:          primitive.NewObjectID(),
		Title:       input.Title,
		Description: input.Description,
		CreatedAt:   time.Now(),
	}

	if err := database.InsertMission(mission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la misión"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Misión creada exitosamente", "mission": mission})
}

// ListMissions - GET /missions (Lista todas las misiones)
func GetAllMissions(c *gin.Context) {
	missions, err := database.GetAllMissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener las misiones"})
		return
	}
	c.JSON(http.StatusOK, missions)
}

// GetMission - GET /missions/:id (Obtiene una misión por su ID)
func GetMissionByID(c *gin.Context) {
	idParam := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de misión inválido"})
		return
	}

	mission, err := database.GetMissionByID(objID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Misión no encontrada"})
		return
	}

	c.JSON(http.StatusOK, mission)
}

func GetMissionsOverview(c *gin.Context) {
	overview, err := database.GetMissionsOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener el overview"})
		return
	}
	c.JSON(http.StatusOK, overview)
}
