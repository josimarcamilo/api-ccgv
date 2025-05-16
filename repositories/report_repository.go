package repositories

import "jc-financas/models"

type AccountBalance struct {
	AccountID uint
	Total     float64
}

func GetBalance(endDate string, teamId uint) []AccountBalance {
	var accountBalances []AccountBalance
	DB.Model(&models.Transaction{}).
		Select("account_id, SUM(CASE WHEN type = 1 THEN value ELSE -value END) as total").
		Where("team_id = ?", teamId).
		Where("date <= ?", endDate).
		Group("account_id").
		Scan(&accountBalances)

	return accountBalances
}
