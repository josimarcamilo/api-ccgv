package repositories

import (
	"jc-financas/models"
)

func CreateUser(model *models.User) error {
	return DB.Create(model).Error
}
