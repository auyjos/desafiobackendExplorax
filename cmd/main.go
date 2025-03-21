package main

import (
	"log"
	"os"

	_ "explorax-backend/docs"
	"explorax-backend/internal/database"
	"explorax-backend/internal/handlers"
	"explorax-backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Explorax Backend API
// @version 1.0
// @description Documentación de la API para Explorax Backend
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró el archivo .env o hubo un error al cargarlo")
	}

	// Conectar a MongoDB
	database.Connect()

	// Configurar Gin Router
	router := gin.Default()
	router.Use(cors.Default())
	// Agregar Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Grupo de endpoints de autenticación
	auth := router.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// Endpoints protegidos con JWT
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware())
	{
		admin.POST("/missions/create", handlers.CreateMission)
	}

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

	// Endpoints públicos
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

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Use(cors.Default())

	router.Run(":" + port)
}
