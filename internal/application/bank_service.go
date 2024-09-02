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

func (s *BankService) Transfer(trf domainBank.TransferTransaction) (uuid.UUID, bool, error) {
	// get from account by account number from
	accountNumberFrom := trf.FromAccountNumber
	accountnumberTo := trf.ToAccountNumber
	if trf.Amount < 0 {
		logErr := util.LogError(fmt.Sprintf("Amount is less than  0 : %v\n", trf.Amount), "", "Bank Service - Transfer - Checking Amount")
		log.Error().Msg(logErr)
		return uuid.Nil, false, domainBank.ErrTransferRecordFailed
	}
	now := time.Now()

	currencySet := map[string]bool{
		"USD": true,
		"IDR": true,
	}

	if !currencySet[trf.Currency] {
		logErr := util.LogError("currency is not available", "", "Bank Service - Transfer - Checking Amount")
		log.Error().Msg(logErr)
		return uuid.Nil, false, domainBank.ErrTransferRecordFailed
	}

	amountTransfer := trf.Amount
	if trf.Currency == "IDR" {
		rate, err := s.db.GetExchangeRateAtTimestamp("USD", "IDR", time.Now())
		if err != nil {
			logErr := util.LogError(fmt.Sprintf("Can't GetExchangeRateAtTimestamp : %v\n", err), "", "Bank Service - Transfer")
			log.Error().Msg(logErr)
			return uuid.Nil, false, domainBank.ErrTransferRecordFailed
		}
		amountTransfer = trf.Amount / rate.Rate
	}

	bankAccountDetailFrom, err := s.db.GetDetailBankAccountByAccountNumber(accountNumberFrom)
	if err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't GetDetailBankAccountByAccountNumber From : %v\n", err), "", "Bank Service - Transfer")
		log.Error().Msg(logErr)
		return uuid.Nil, false, domainBank.ErrTransferSourceAccountNotFound
	}

	if bankAccountDetailFrom.CurrentBalance < amountTransfer {
		return uuid.Nil, false, domainBank.ErrTransferTransactionPair
	}

	bankAccountDetailTo, err := s.db.GetDetailBankAccountByAccountNumber(accountnumberTo)
	if err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't GetDetailBankAccountByAccountNumber To : %v\n", err), "", "Bank Service - Transfer")
		log.Error().Msg(logErr)
		return uuid.Nil, false, domainBank.ErrTransferDestinationAccountNotFound
	}

	transferDetail := domainBank.BankTransferOrm{
		TransferUuid:      uuid.New(),
		FromAccountUuid:   bankAccountDetailFrom.AccountUuid,
		ToAccountUuid:     bankAccountDetailTo.AccountUuid,
		Currency:          trf.Currency,
		Amount:            amountTransfer,
		TransferTimestamp: now,
		TransferSuccess:   false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	uuidTrans, err := s.db.CreateTransfer(transferDetail)
	if err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't CreateTransfer : %v\n", err), "", "Bank Service - Transfer")
		log.Error().Msg(logErr)
		return uuid.Nil, false, domainBank.ErrTransferRecordFailed
	}

	bankTransactionOrmFrom := domainBank.BankTransactionOrm{
		TransactionUuid:      uuid.New(),
		AccountUuid:          bankAccountDetailFrom.AccountUuid,
		TransactionTimestamp: now,
		Amount:               amountTransfer,
		TransactionType:      "1",
		Notes:                trf.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	bankTransactionOrmTo := domainBank.BankTransactionOrm{
		TransactionUuid:      uuid.New(),
		AccountUuid:          bankAccountDetailFrom.AccountUuid,
		TransactionTimestamp: now,
		Amount:               amountTransfer,
		TransactionType:      "2",
		Notes:                trf.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	status, err := s.db.CreateTransferTransactionPair(bankAccountDetailFrom, bankAccountDetailTo, bankTransactionOrmFrom, bankTransactionOrmTo)
	if err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't CreateTransferTransactionPair : %v\n", err), "", "Bank Service - Transfer")
		log.Error().Msg(logErr)
		return uuid.Nil, false, domainBank.ErrTransferTransactionPair
	}

	err = s.db.UpdateTransferStatus(transferDetail, status)
	if err != nil {
		return uuid.Nil, false, domainBank.ErrTransferRecordFailed
	}

	return uuidTrans, true, nil

}
