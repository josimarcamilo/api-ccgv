package controllers

import (
	"jc-financas/models"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func RegisterUser(c echo.Context) error {
	var user models.User

	if err := c.Bind(&user); err != nil {
		return errors.Wrap(err, "bind request")
	}

	c.JSON(http.StatusOK, user)
	return nil

	// if err := c.ShouldBindJSON(&user); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	// if result := database.DB.Create(&user); result.Error != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, user)
}

// func ShowRegisterPage(c echo.Context) error {
// 	data := map[string]interface{}{
// 		"Title": "Register Page",
// 	}

// 	return c.HTML(http.StatusOK, "register.html", data)
// }
