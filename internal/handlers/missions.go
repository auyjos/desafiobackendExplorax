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

type GenericResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"` // Si hay error, lo incluye
}

type LeaderboardEntry struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	CompletedCount int    `json:"completed_count"`
}

// UserStatistics contiene estadísticas generales del usuario
type UserStatistics struct {
	TotalCompleted     int     `json:"total_completed"`
	AverageDuration    float64 `json:"average_duration"`
	ProgressPercentage float64 `json:"progress_percentage"`
}

// MissionsOverview contiene el resumen de misiones
type MissionsOverview struct {
	MostPopularMission map[string]interface{}   `json:"most_popular_mission"`
	AvgCompletionTimes []map[string]interface{} `json:"avg_completion_times"`
}

// StartMissionRequest representa el cuerpo de la solicitud para iniciar una misión.
// @Description Estructura del request para iniciar misión
type StartMissionRequest struct {
	MissionID string `json:"mission_id" binding:"required" example:"60a7b97f5e41c42e7c2e30b6"`
}

// StartMission godoc
// @Summary Inicia una misión
// @Description Registra el progreso de una misión como "iniciada" para el usuario autenticado.
// @Tags Missions
// @Accept json
// @Produce json
// @Param Authorization header string true "Token de autorización (Bearer {token})"
// @Param mission body StartMissionRequest true "ID de la misión a iniciar"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Misión iniciada exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "Usuario no autenticado"
// @Failure 500 {object} map[string]string "Error al iniciar la misión"
// @Router /missions/start [post]

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

	// Vincular el JSON recibido a la estructura StartMissionRequest
	var input StartMissionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Convertir MissionID a ObjectID
	missionObjID, err := primitive.ObjectIDFromHex(input.MissionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de misión inválido"})
		return
	}

	progress := models.MissionProgress{
		ID:        primitive.NewObjectID(),
		UserID:    userObjID,
		MissionID: missionObjID,
		Status:    "iniciada",
		StartDate: time.Now(),
	}

	if err := database.InsertMissionProgress(progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al iniciar la misión"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Misión iniciada"})
}

// CompleteMission godoc
// @Summary Completa una misión
// @Description Actualiza el estado de la misión a "completada".
// @Tags Missions
// @Accept  json
// @Produce  json
// @Param mission body object true "Mission ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /missions/complete [post]
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

// GetProgress godoc
// @Summary Obtiene el progreso de misiones
// @Description Devuelve todas las misiones iniciadas y completadas del usuario.
// @Tags Missions
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} []models.MissionProgress
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /missions/progress [get]
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

// GetActiveMissions godoc
// @Summary Obtiene misiones activas de un usuario
// @Description Retorna todas las misiones con estado "iniciada" para un usuario autenticado
// @Tags Missions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.MissionProgress
// @Failure 401 {object} GenericResponse "Usuario no autenticado"
// @Failure 500 {object} GenericResponse "Error interno del servidor"
// @Router /missions/active [get]
func GetActiveMissions(c *gin.Context) {
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

// GetCompletedMissions godoc
// @Summary Obtiene misiones completadas de un usuario
// @Description Retorna todas las misiones con estado "completada" para un usuario autenticado
// @Tags Missions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.MissionProgress
// @Failure 401 {object} GenericResponse "Usuario no autenticado"
// @Failure 500 {object} GenericResponse "Error interno del servidor"
// @Router /missions/completed [get]
func GetCompletedMissions(c *gin.Context) {
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

// GetLeaderboard godoc
// @Summary Obtiene el ranking de usuarios basado en misiones completadas
// @Description Devuelve un leaderboard con los usuarios ordenados por número de misiones completadas
// @Tags Missions
// @Accept json
// @Produce json
// @Success 200 {array} LeaderboardEntry
// @Failure 500 {object} GenericResponse "Error interno del servidor"
// @Router /missions/leaderboard [get]
func GetLeaderboard(c *gin.Context) {
	leaderboard, err := database.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el leaderboard"})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

// GetStatistics godoc
// @Summary Obtiene estadísticas del usuario
// @Description Devuelve estadísticas como el número total de misiones completadas y duración promedio
// @Tags Missions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserStatistics
// @Failure 401 {object} GenericResponse "Usuario no autenticado"
// @Failure 500 {object} GenericResponse "Error interno del servidor"
// @Router /missions/statistics [get]
func GetStatistics(c *gin.Context) {
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

	stats, err := database.GetUserStatistics(userObjID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener estadísticas"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateMission godoc
// @Summary Crea una nueva misión
// @Description Crea una nueva misión con un título y descripción proporcionados
// @Tags Missions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param mission body models.Mission true "Detalles de la misión"
// @Success 201 {object} GenericResponse "Misión creada exitosamente"
// @Failure 400 {object} GenericResponse "Datos inválidos"
// @Failure 500 {object} GenericResponse"No se pudo crear la misión"
// @Router /missions [post]
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

// GetAllMissions godoc
// @Summary Obtiene todas las misiones
// @Description Retorna una lista de todas las misiones disponibles en la base de datos
// @Tags Missions
// @Accept json
// @Produce json
// @Success 200 {array} models.Mission
// @Failure 500 {object} GenericResponse "No se pudieron obtener las misiones"
// @Router /missions [get]
func GetAllMissions(c *gin.Context) {
	missions, err := database.GetAllMissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron obtener las misiones"})
		return
	}
	c.JSON(http.StatusOK, missions)
}

// GetMissionByID godoc
// @Summary Obtiene una misión por su ID
// @Description Retorna los detalles de una misión específica por su ID
// @Tags Missions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID de la misión"
// @Success 200 {object} models.Mission
// @Failure 400 {object} GenericResponse "ID de misión inválido"
// @Failure 404 {object}  GenericResponse "Misión no encontrada"
// @Router /mission/{id} [get]
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

// GetMissionsOverview obtiene una visión general de las misiones.
//
// @Summary Obtiene el resumen de las misiones
// @Description Devuelve estadísticas sobre las misiones, incluyendo la misión más popular y el tiempo promedio de finalización.
// @Tags Missions
// @Produce json
// @Success 200 {object} map[string]interface{} "Resumen de misiones"
// @Failure 500 {object} map[string]interface{} "Error interno al obtener el resumen"
// @Router /missions/overview [get]
func GetMissionsOverview(c *gin.Context) {
	overview, err := database.GetMissionsOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo obtener el overview"})
		return
	}
	c.JSON(http.StatusOK, overview)
}
