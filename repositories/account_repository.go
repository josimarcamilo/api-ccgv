package repositories

import (
	"jc-financas/models"
)

func CreateAccount(model *models.Account) error {
	return DB.Create(model).Error
}

func GetAccount(unidadeId, teamID uint) (*models.Account, error) {
	var model models.Account
	if err := DB.Where("team_id = ?", teamID).Where("id = ?", unidadeId).Preload("Unidade").First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func UpdateAccount(model *models.Account) error {
	return DB.Save(&model).Error
}

func GetAccounts(teamID uint) ([]models.Account, error) {
	var models []models.Account

	if err := DB.Where("team_id = ?", teamID).Preload("Unidade").Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
