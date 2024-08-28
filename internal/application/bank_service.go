package application

import (
	"github.com/fajaramaulana/go-grpc-micro-bank-server/internal/port"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/rs/zerolog/log"
)

type BankService struct {
	db port.BankDatabasePort
}

func NewBankService(dbPort port.BankDatabasePort) *BankService {
	return &BankService{
		db: dbPort,
	}
}

func (s *BankService) GetCurrentBalance(account string) (float64, error) {
	bankAccount, err := s.db.GetBalanceBankAccountByAccountNumber(account)

	if err != nil {
		logErr := util.LogError("Error on FindCurrentBalance: "+err.Error(), "", "DatabaseAdapter - GetBankAccountByAccountNumber")
		log.Error().Msg(logErr)
		return 0, err
	}

	return bankAccount.CurrentBalance, nil
}
