package repositories

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
