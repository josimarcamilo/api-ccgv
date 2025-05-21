package repositories

import "jc-financas/models"

type AccountBalanceReport struct {
	Accounts        []AccountBalance
	VirtualAccounts []VirtualAccountBalance
	Balance         int
}
type AccountBalance struct {
	AccountID uint
	Total     int
}
type VirtualAccountBalance struct {
	AccountVirtualID uint
	Total            int
}

func GetBalance(endDate string, teamId uint) AccountBalanceReport {
	var accountBalances []AccountBalance
	var virtualAccountBalances []VirtualAccountBalance
	DB.Model(&models.Transaction{}).
		Select("account_id, SUM(CASE WHEN type = 1 THEN value ELSE -value END) as total").
		Where("team_id = ?", teamId).
		Where("account_id is not null").
		Where("date <= ?", endDate).
		Group("account_id").
		Scan(&accountBalances)

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
