package application

import (
	"time"

	domainBank "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application/domain/bank"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/internal/port"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/google/uuid"
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

func (s *BankService) CreateExchangeRate(r domainBank.ExchangeRate) (uuid.UUID, error) {
	newUuid := uuid.New()
	now := time.Now()

	exchangeRateOrm := domainBank.BankExchangeRateOrm{
		ExchangeRateUuid:   newUuid,
		FromCurrency:       r.FromCurrency,
		ToCurrency:         r.ToCurrency,
		Rate:               r.Rate,
		ValidFromTimestamp: r.ValidFromTimestamp,
		ValidToTimestamp:   r.ValidToTimestamp,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	return s.db.InsertExchangeRate(exchangeRateOrm)
}

func (s *BankService) FindExchangeRate(fromCurrency string, toCurrency string, ts time.Time) (float64, error) {
	exchangeRate, err := s.db.GetExchangeRateAtTimestamp(fromCurrency, toCurrency, ts)

	if err != nil {
		logErr := util.LogError("Error on FindExchangeRate: "+err.Error(), "", "DatabaseAdapter - GetExchangeRateAtTimestamp")
		log.Error().Msg(logErr)

		return 0, err
	}

	return exchangeRate.Rate, nil
}
