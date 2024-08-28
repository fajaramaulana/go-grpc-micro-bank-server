package port

type BankServicePort interface {
	GetCurrentBalance(account string) (float64, error)
}
