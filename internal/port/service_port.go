package port

import (
	"time"

	domainBank "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application/domain/bank"
	"github.com/google/uuid"
)

type BankServicePort interface {
	GetCurrentBalance(account string) (float64, error)
	CreateExchangeRate(r domainBank.ExchangeRate) (uuid.UUID, error)
	FindExchangeRate(fromCurrency string, toCurrency string, ts time.Time) (float64, error)
	CreateTransaction(accountNum string, trx domainBank.Transaction) (uuid.UUID, error)
	CalculateTransactionSummary(trxSum *domainBank.TransactionSummary, trx domainBank.Transaction) error
	Transfer(trf domainBank.TransferTransaction) (uuid.UUID, bool, error)
}
