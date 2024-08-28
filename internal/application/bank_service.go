package application

type BankService struct {
}

func (a *BankService) GetCurrentBalance(account string) (float64, error) {
	return 100000, nil
}
