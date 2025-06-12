package controllers

import (
	"fmt"
	"jc-financas/models"
	"jc-financas/repositories"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func CreateTransaction(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var model models.Transaction
	if err := c.Bind(&model); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Erro ao interpretar a requisição",
			"message": err.Error(),
		})
	}

	model.TeamID = claims.TeamID

	// validações obrigatórias
	if model.Type == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Campo 'type' é obrigatório",
		})
	}
	if model.Date == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Campo 'date' é obrigatório",
		})
	}

	// valida conta (se enviada)
	if model.AccountID != nil {
		account, err := repositories.GetAccount(*model.AccountID, claims.TeamID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "Conta não encontrada",
				"message": err.Error(),
			})
		}
		if account.TeamID != claims.TeamID {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Conta não pertence à equipe",
			})
		}
	}

	if model.AccountVirtualID != nil {
		account, err := repositories.GetAccount(*model.AccountVirtualID, claims.TeamID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "Conta virtual não encontrada",
				"message": err.Error(),
			})
		}
		if account.TeamID != claims.TeamID {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Conta virtual não pertence à equipe",
			})
		}
	}

	// valida categoria (se enviada)
	if model.CategoryID != nil {
		category, err := repositories.GetCategory(*model.CategoryID, claims.TeamID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "Categoria não encontrada",
				"message": err.Error(),
			})
		}
		if category.TeamID != claims.TeamID {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Categoria não pertence à equipe",
			})
		}
	}

	// valida categoryMap (se enviada)
	if model.CategoryMapID != nil {
		categoryMap, err := repositories.GetCategory(*model.CategoryMapID, claims.TeamID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":   "Categoria do mapa não encontrada",
				"message": err.Error(),
			})
		}
		if categoryMap.TeamID != claims.TeamID {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Categoria do mapa não pertence à equipe",
			})
		}
	}

	// cria transação
	if err := repositories.CreateTransaction(&model); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao criar transação",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, model)
}

func GetTransaction(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var search models.Transaction
	// bind
	if err := c.Bind(&search); err != nil {
		return errors.Wrap(err, "bind request")
	}

	// validacao
	if search.ID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "transacao é obrigatória",
		})
	}

	find, err := repositories.GetTransaction(search.ID, claims.TeamID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "transacao não encontrada",
			"message": err.Error(),
		})
	}

	c.JSON(http.StatusOK, find)
	return nil
}

func UpdateTransaction(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var tx models.Transaction
	if err := c.Bind(&tx); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Erro ao interpretar o corpo da requisição",
			"message": err.Error(),
		})
	}

	// ID obrigatório para saber o que atualizar
	if tx.ID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID da transação é obrigatório",
		})
	}

	// Busca original para validar team
	fmt.Println(tx.ID, claims.TeamID)
	existing, err := repositories.GetTransaction(tx.ID, claims.TeamID)
	if err != nil || existing.TeamID != claims.TeamID {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Transação não encontrada ou não pertence à equipe",
		})
	}

	existing.AccountID = tx.AccountID
	existing.AccountVirtualID = tx.AccountVirtualID
	existing.CategoryID = tx.CategoryID
	existing.CategoryMapID = tx.CategoryMapID
	existing.Type = tx.Type
	existing.IsTransfer = tx.IsTransfer
	existing.Date = tx.Date
	existing.Description = tx.Description
	existing.Value = tx.Value

	// Atualiza
	if err := repositories.UpdateTransaction(existing); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao atualizar transação",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Transação atualizada com sucesso",
		"id":      tx.ID,
	})
}

func Delete(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var tx models.Transaction
	if err := c.Bind(&tx); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Erro ao interpretar o corpo da requisição",
			"message": err.Error(),
		})
	}

	// ID obrigatório para saber o que atualizar
	if tx.ID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "ID da transação é obrigatório",
		})
	}

	// Busca original para validar team
	fmt.Println(tx.ID, claims.TeamID)
	existing, err := repositories.GetTransaction(tx.ID, claims.TeamID)
	if err != nil || existing.TeamID != claims.TeamID {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Transação não encontrada ou não pertence à equipe",
		})
	}

	if err := repositories.DeleteTransaction(existing); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao deletar transação",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Transação deletada com sucesso",
	})
}

func ListTransactions(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	var filer models.TransactionFilter
	if err := c.Bind(&filer); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "Binf error",
			"message": err.Error(),
		})
	}

	results, err := repositories.GetTransactions(claims.TeamID, filer)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   "Erro ao buscar contas",
			"message": err.Error(),
		})
	}
	// vamos tratar o resultado para pegar apenas o que queremos de results
	var rows []models.TransactionList
	for _, item := range results {
		rows = append(rows, models.TransactionList{
			ID:                 item.ID,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
			Type:               item.Type,
			IsTransfer:         item.IsTransfer,
			Date:               item.Date,
			Description:        item.Description,
			Value:              item.Value,
			ReceiptUrl:         item.ReceiptUrl,
			AccountID:          item.Account.ID,
			AccountName:        item.Account.Name,
			AccountVirtualID:   item.AccountVirtual.ID,
			AccountVirtualName: item.AccountVirtual.Name,
			CategoryID:         item.Category.ID,
			CategoryName:       item.Category.Name,
			CategoryMapID:      item.CategoryMap.ID,
			CategoryMapName:    item.CategoryMap.Name,
		})
	}

	c.JSON(http.StatusOK, rows)
	return nil
}
