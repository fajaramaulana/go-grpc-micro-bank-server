package application

import (
	"fmt"
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

func (s *BankService) CreateTransaction(accountNum string, trx domainBank.Transaction) (uuid.UUID, error) {
	newUuid := uuid.New()
	now := time.Now()

	bankAccountDetail, err := s.db.GetDetailBankAccountByAccountNumber(accountNum)

	if err != nil {
		logErr := util.LogError("Error on GetDetailBankAccountByAccountNumber: "+err.Error(), "", "Bank Service - CreateTransaction")
		log.Error().Msg(logErr)
		return uuid.Nil, err
	}

	// Check if the transaction is an "out" transaction and if the account has sufficient balance
	if trx.TransactionType == domainBank.TransactionTypeOut && bankAccountDetail.CurrentBalance < trx.Amount {
		err := fmt.Errorf("insufficient balance: transaction amount %v exceeds current balance %v", trx.Amount, bankAccountDetail.CurrentBalance)
		logErr := util.LogError(fmt.Sprintf("Can't create transaction : %v\n", err), "", "BankAdapter - CreateTransaction")
		log.Error().Msg(logErr)
		return uuid.Nil, err
	}

	transactionOrm := domainBank.BankTransactionOrm{
		TransactionUuid:      newUuid,
		AccountUuid:          bankAccountDetail.AccountUuid,
		TransactionTimestamp: now,
		Amount:               trx.Amount,
		TransactionType:      trx.TransactionType,
		Notes:                trx.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	saveUuid, err := s.db.CreateTransaction(bankAccountDetail, transactionOrm)
	if err != nil {
		logErr := util.LogError("Error on CreateTransaction: "+err.Error(), "", "Bank Service - CreateTransaction")
		log.Error().Msg(logErr)
		return uuid.Nil, err
	}

	return saveUuid, nil
}

func (s *BankService) CalculateTransactionSummary(trxSum *domainBank.TransactionSummary, trx domainBank.Transaction) error {
	switch trx.TransactionType {
	case domainBank.TransactionTypeIn:
		trxSum.SumIn += trx.Amount
	case domainBank.TransactionTypeOut:
		trxSum.SumOut += trx.Amount
	default:
		return fmt.Errorf("unknown transaction type %v", trx.TransactionType)
	}

	trxSum.SumTotal = trxSum.SumIn - trxSum.SumOut

	return nil
}
