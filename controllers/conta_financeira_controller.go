package controllers

import (
	"jc-financas/models"
	"jc-financas/repositories"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func CreateAccount(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var model models.Account
	// bind
	if err := c.Bind(&model); err != nil {
		return errors.Wrap(err, "bind request")
	}

	model.TeamID = claims.TeamID

	// validacao
	if model.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Nome obrigatório",
		})
	}
	if model.UnidadeID != nil {
		unidade, err := repositories.GetUnidade(*model.UnidadeID, claims.TeamID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "Unidade não encontrada",
				"message": err.Error(),
			})
		}

		if unidade.TeamID != claims.TeamID {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Unidade não pertence a equipe",
			})
		}
	}

	if err := repositories.CreateAccount(&model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao criar conta",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusCreated, model)
	return nil
}

func GetAccount(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var search models.Account
	// bind
	if err := c.Bind(&search); err != nil {
		return errors.Wrap(err, "bind request")
	}

	// validacao
	if search.ID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Conta obrigatória",
		})
	}

	find, err := repositories.GetAccount(search.ID, claims.TeamID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Conta não encontrada",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, find)
	return nil
}

func UpdadeAccount(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var search models.Account
	// bind
	if err := c.Bind(&search); err != nil {
		return errors.Wrap(err, "bind request")
	}

	// validacao
	if search.ID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Conta obrigatória",
		})
	}

	find, err := repositories.GetAccount(search.ID, claims.TeamID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Conta não encontrada",
			"message": err.Error(),
		})
	}

	if search.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Nome obrigatório",
		})
	}

	if search.UnidadeID != nil {
		unidade, err := repositories.GetUnidade(*search.UnidadeID, claims.TeamID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "Unidade não encontrada",
				"message": err.Error(),
			})
		}

		if unidade.TeamID != claims.TeamID {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Unidade não pertence a equipe",
			})
		}
		find.UnidadeID = search.UnidadeID
	}

	find.Name = search.Name
	if err := repositories.UpdateAccount(find); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao atualizar conta",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, find)
	return nil
}

func ListAccounts(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	results, err := repositories.GetAccounts(claims.TeamID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar contas",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, results)
	return nil
}
