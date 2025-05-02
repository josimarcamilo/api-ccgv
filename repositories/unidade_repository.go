package repositories

import (
	"jc-financas/models"
)

func CreateUnidade(model *models.Unidade) error {
	return DB.Create(model).Error
}

func GetUnidade(model *models.Unidade, teamID uint) error {
	return DB.Where("team_id = ?", teamID).First(&model).Error
}

func GetUnidades(teamID uint) ([]models.Unidade, error) {
	var models []models.Unidade

	if err := DB.Where("team_id = ?", teamID).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
