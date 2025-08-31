package controllers

import (
	"jc-financas/repositories"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Report struct {
	AccountID uint   `param:"account"`
	EndDate   string `query:"end_date" json:"end_date"`
}

type ExtractReport struct {
	StartDate string `query:"start_date" json:"start_date"`
	EndDate   string `query:"end_date" json:"end_date"`
}

func GetBalance(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}
	var report Report
	if err := c.Bind(&report); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "bind",
			"message": err.Error(),
		})
	}

	if report.EndDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Campo 'end_date' é obrigatório",
		})
	}

	return c.JSON(http.StatusOK, repositories.GetAllBalances(report.EndDate, claims.TeamID))
}

func GetExtact(c echo.Context) error {
	claims, err := repositories.ParseWithContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}
	var report ExtractReport
	if err := c.Bind(&report); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":   "bind",
			"message": err.Error(),
		})
	}

	if report.StartDate == "" || report.EndDate == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Campo 'start_date' e 'end_date' são obrigatórios",
		})
	}

	return c.JSON(http.StatusOK, repositories.GetExtract(report.StartDate, report.EndDate, claims.TeamID))
}
