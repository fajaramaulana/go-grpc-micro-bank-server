package port

import (
	"time"

	domainBank "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application/domain/bank"
	"github.com/google/uuid"
)

type BankDatabasePort interface {
	GetBalanceBankAccountByAccountNumber(acct string) (domainBank.BalanceAccountOrm, error)
	InsertExchangeRate(r domainBank.BankExchangeRateOrm) (uuid.UUID, error)
	GetExchangeRateAtTimestamp(fromCurrency string, toCurrency string, ts time.Time) (domainBank.BankExchangeRateOrm, error)
}
