package main

import (
	"log"
	"os"

	"explorax-backend/internal/database"
	"explorax-backend/internal/handlers"
	"explorax-backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carga las variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró el archivo .env o hubo un error al cargarlo")
	}

	// Conecta a MongoDB
	database.Connect()

	// Crea el router
	router := gin.Default()

	// Grupo de endpoints de autenticación
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware())
	{
		admin.POST("/missions/create", handlers.CreateMission)
	}
	// Grupo de endpoints de misiones
	missions := router.Group("/missions")
	missions.Use(middleware.JWTAuthMiddleware())
	{

		missions.GET("/all", handlers.GetAllMissions)
		missions.POST("/start", handlers.StartMission)
		missions.POST("/complete", handlers.CompleteMission)
		missions.GET("/progress", handlers.GetProgress)
		missions.GET("/active", handlers.GetActiveMissions)
		missions.GET("/completed", handlers.GetCompletedMissions)
		missions.GET("/statistics", handlers.GetStatistics)
	}

	publicMissions := router.Group("/missions")
	{
		publicMissions.GET("/leaderboard", handlers.GetLeaderboard)
		publicMissions.GET("/overview", handlers.GetMissionsOverview)
	}

	mission := router.Group("/mission")
	mission.Use(middleware.JWTAuthMiddleware())
	{
		mission.GET("/:id", handlers.GetMissionByID)
	}

	// Define el puerto (por defecto 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Inicia el servidor
	router.Run(":" + port)
}
