// /internal/handlers/missions.go
package handlers

import (
	"net/http"
	"time"

	"explorax-backend/internal/database"
	"explorax-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// StartMission - POST /missions/start
// Inicia una misión registrando el user_id (del JWT), mission_id, fecha de inicio y estado "iniciada".
func StartMission(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	var input struct {
		MissionID string `json:"mission_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	progress := models.MissionProgress{
		UserID:    userID.(string),
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}
	var input struct {
		MissionID string `json:"mission_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	if err := database.UpdateMissionProgress(userID.(string), input.MissionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al completar la misión"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Misión completada"})
}

// GetProgress - GET /missions/progress
// Devuelve el progreso de misiones (tanto iniciadas como completadas) del usuario.
func GetProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	progress, err := database.GetMissionProgress(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el progreso"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetActiveMissions - GET /missions/active
// Retorna las misiones iniciadas que aún no se han completado.
func GetActiveMissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	active, err := database.GetActiveMissions(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener misiones activas"})
		return
	}

	c.JSON(http.StatusOK, active)
}

// GetCompletedMissions - GET /missions/completed
// Retorna las misiones completadas junto con la fecha de finalización.
func GetCompletedMissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	completed, err := database.GetCompletedMissions(userID.(string))
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	stats, err := database.GetUserStatistics(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener estadísticas"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
