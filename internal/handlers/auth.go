package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"explorax-backend/internal/database"
	"explorax-backend/internal/models"
	"explorax-backend/internal/utils"
)

// RegisterRequest representa los datos esperados en el registro de usuario.
// @Description Estructura para registrar un usuario
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"usuario123"`
	Email    string `json:"email" binding:"required,email" example:"usuario@email.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// LoginRequest representa los datos esperados en el login de usuario.
// @Description Estructura para iniciar sesión
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"usuario@email.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// Register godoc
// @Summary Registro de usuario
// @Description Permite registrar un nuevo usuario en la plataforma.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Datos del usuario"
// @Success 201 {object} map[string]string "Usuario creado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 500 {object} map[string]string "Error interno del servidor"
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var input RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Encriptar la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al encriptar la contraseña"})
		return
	}

	user := models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}

	// Insertar el usuario en MongoDB
	if err := database.InsertUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al registrar el usuario"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuario creado exitosamente"})
}

// Login godoc
// @Summary Inicia sesión de usuario
// @Description Autentica a un usuario y devuelve un token JWT.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param credentials body LoginRequest true "Credenciales de usuario (email y password)"
// @Success 200 {object} map[string]string "Token JWT generado exitosamente"
// @Failure 400 {object} map[string]string "Datos inválidos"
// @Failure 401 {object} map[string]string "Credenciales incorrectas"
// @Failure 500 {object} map[string]string "Error interno"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var input LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// Buscar usuario por email
	user, err := database.FindUserByEmail(input.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Comparar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Contraseña incorrecta"})
		return
	}

	// Generar token JWT
	token, err := utils.GenerateJWT(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
