package repositories

import (
	"jc-financas/models"
)

func CreateUserTeam(userId, teamId, RoleId uint) error {
	return DB.Create(&models.UserTeam{
		UserID: userId,
		TeamID: teamId,
		RoleId: RoleId,
	}).Error
}
