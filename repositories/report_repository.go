package repositories

import "jc-financas/models"

type AccountBalanceReport struct {
	Accounts        []AccountBalance
	VirtualAccounts []VirtualAccountBalance
	Balance         int
}
type AccountBalance struct {
	AccountName string `gorm:"column:name"` // Alias usado no SELECT
	AccountID   uint   `gorm:"column:account_id"`
	Total       int    `gorm:"column:total"`
}
type VirtualAccountBalance struct {
	AccountName      string `gorm:"column:name"` // Alias usado no SELECT
	AccountVirtualID uint   `gorm:"column:account_virtual_id"`
	Total            int    `gorm:"column:total"`
}

func GetBalance(endDate string, teamId uint) AccountBalanceReport {
	var accountBalances []AccountBalance
	DB.Model(&models.Transaction{}).
		Joins("JOIN accounts ON accounts.id = transactions.account_id").
		Select("accounts.name AS name, account_id, SUM(CASE WHEN type = 1 THEN value ELSE -value END) AS total").
		Where("transactions.team_id = ?", teamId).
		Where("account_id IS NOT NULL").
		Where("date <= ?", endDate).
		Group("accounts.name, account_id").
		Order("total desc").
		Scan(&accountBalances)

	var virtualAccountBalances []VirtualAccountBalance
	DB.Model(&models.Transaction{}).
		Select("account_virtual_id, SUM(CASE WHEN type = 1 THEN value ELSE -value END) as total").
		Where("team_id = ?", teamId).
		Where("account_virtual_id is not null").
		Where("date <= ?", endDate).
		Group("account_virtual_id").
		Scan(&virtualAccountBalances)

	// Soma dos saldos de contas reais
	var realTotal int
	for _, acc := range accountBalances {
		realTotal += acc.Total
	}

	// Soma dos saldos de contas virtuais
	var virtualTotal int
	for _, vAcc := range virtualAccountBalances {
		virtualTotal += vAcc.Total
	}

	// Balance = total das reais - total das virtuais
	finalBalance := realTotal - virtualTotal

	return AccountBalanceReport{
		Accounts:        accountBalances,
		VirtualAccounts: virtualAccountBalances,
		Balance:         finalBalance,
	}
}
