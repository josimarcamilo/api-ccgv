package repositories

type AccountBalanceReport struct {
	Accounts        []AccountBalance
	VirtualAccounts []AccountBalance
	Balance         int
}
type AccountBalance struct {
	AccountName string `gorm:"column:name"`       // Alias usado no SELECT
	AccountID   uint   `gorm:"column:account_id"` // Alias usado no SELECT
	Balance     int    `gorm:"column:balance"`    // Alias usado no SELECT
}
