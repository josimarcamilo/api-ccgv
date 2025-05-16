package controllers

import (
	"jc-financas/repositories"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Report struct {
	EndDate string `query:"end_date" json:"end_date"`
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

	return c.JSON(http.StatusOK, repositories.GetBalance(report.EndDate, claims.TeamID))
}
