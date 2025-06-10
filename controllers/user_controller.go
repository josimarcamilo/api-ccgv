package controllers

import (
	"fmt"
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
	if err := repositories.DB.Where("email = ?", loginRequest.Email).Preload("Team").First(&user).Error; err != nil {
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

func AddUserToTeam(c echo.Context) error {

	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	// return c.JSON(http.StatusUnauthorized, claims)

	type LoginRequest struct {
		Name  string `json:"name" form:"name"`
		Email string `json:"email" form:"email"`
	}

	var loginRequest LoginRequest
	if err := c.Bind(&loginRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Dados inválidos",
		})
	}

	if loginRequest.Email == "" || loginRequest.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email e nome obrigatório",
		})
	}

	newPassword, err := repositories.GeneratePassword(10)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Erro ao gerar senha",
		})
	}

	// Buscar o usuário pelo e-mail
	var user models.User
	repositories.DB.Where("email = ?", loginRequest.Email).First(&user)

	if user.ID != 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Usuário já existe",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Erro ao encriptar senha",
		})
	}

	user.Password = string(hashedPassword)
	user.Name = loginRequest.Name
	user.Email = loginRequest.Email
	user.TeamID = &claims.TeamID

	if err := repositories.CreateUser(&user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao criar usuário",
			"message": err.Error(),
		})
	}

	// role 1 sudo, 2 usuario
	err = repositories.CreateUserTeam(user.ID, claims.TeamID, 2)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao adicionar usuário ao time",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"email":    user.Email,
		"password": newPassword,
	})
}

func ListUsersToTeam(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	query := repositories.DB.Table("users").
		Where("team_id = ?", claims.TeamID)

	columnOrder := c.QueryParam("column_order")
	if columnOrder == "" {
		columnOrder = "id"
	}

	columnSort := c.QueryParam("column_sort")
	if columnSort == "" {
		columnSort = "ASC"
	}

	type UserList struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		TeamID    uint   `json:"team_id"`
		CreatedAt string `json:"created_at"`
	}

	var records []UserList
	order := fmt.Sprintf("%s %s", columnOrder, columnSort)
	if err := query.Order(order).Find(&records).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar registros",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, records)
}

func Profile(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, claims)
}
