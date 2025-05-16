package repositories

import (
	"jc-financas/models"
)

func CreateAccount(model *models.Account) error {
	return DB.Create(model).Error
}

func GetAccount(id, teamID uint) (*models.Account, error) {
	var model models.Account
	if err := DB.Where("team_id = ?", teamID).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func UpdateAccount(model *models.Account) error {
	return DB.Save(&model).Error
}

func GetAccounts(teamID uint) ([]models.Account, error) {
	var models []models.Account

	if err := DB.Where("team_id = ?", teamID).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
