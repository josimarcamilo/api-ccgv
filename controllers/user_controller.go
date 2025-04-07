package controllers

import (
	"jc-financas/models"
	"jc-financas/repositories"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(c echo.Context) error {
	var user models.User

	if err := c.Bind(&user); err != nil {
		return errors.Wrap(err, "bind request")
	}

	c.JSON(http.StatusOK, user)
	return nil
}

func Login(c echo.Context) error {
	type LoginRequest struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	var loginRequest LoginRequest
	if err := c.Bind(&loginRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Dados inválidos"})
	}

	// Verificar se o e-mail e senha foram fornecidos
	if loginRequest.Email == "" || loginRequest.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email e senha são obrigatórios"})
	}

	// Buscar o usuário pelo e-mail
	var user models.User
	if err := repositories.DB.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciais inválidas"})
	}

	// Comparar a senha fornecida com o hash armazenado
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciais inválidas"})
	}

	// gere um toke jwt
	token, err := repositories.GenerateJWT(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao gerar token",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}
