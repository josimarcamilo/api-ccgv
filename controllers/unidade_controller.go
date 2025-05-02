package controllers

import (
	"jc-financas/models"
	"jc-financas/repositories"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func CreateUnidade(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var model models.Unidade
	// bind
	if err := c.Bind(&model); err != nil {
		return errors.Wrap(err, "bind request")
	}

	model.TeamID = claims.TeamID

	// validacao
	if model.Nome == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Nome obrigatório",
		})
	}

	if err := repositories.CreateUnidade(&model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao criar unidade",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, model)
	return nil
}

func GetUnidade(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var search models.Unidade
	// bind
	if err := c.Bind(&search); err != nil {
		return errors.Wrap(err, "bind request")
	}

	// validacao
	if search.ID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Unidade obrigatório",
		})
	}

	// search.TeamID = claims.TeamID

	if err := repositories.GetUnidade(&search, claims.TeamID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar unidade",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, search)
	return nil
}

func ListUnidades(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	results, err := repositories.GetUnidades(claims.TeamID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar unidades",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, results)
	return nil
}
