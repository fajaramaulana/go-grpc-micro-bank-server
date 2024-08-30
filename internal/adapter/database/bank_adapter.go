package database

import (
	"fmt"
	"time"

	domainBank "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application/domain/bank"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (a *DatabaseAdapter) GetDetailBankAccountByAccountNumber(accountNum string) (domainBank.BankAccountOrm, error) {
	var bankAccountOrm domainBank.BankAccountOrm

	if err := a.db.First(&bankAccountOrm, "account_number = ?", accountNum).Error; err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't find bank account number %v : %v\n", accountNum, err), "", "BankAdapter - GetDetailBankAccountByAccountNumber")
		log.Error().Msg(logErr)
		return bankAccountOrm, err
	}

	return bankAccountOrm, nil
}

func (a *DatabaseAdapter) GetBalanceBankAccountByAccountNumber(acct string) (domainBank.BalanceAccountOrm, error) {
	var bankAccountOrm domainBank.BalanceAccountOrm

	if err := a.db.Select("account_uuid, account_number, currency, current_balance").First(&bankAccountOrm, "account_number = ?", acct).Error; err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't find bank account number %v : %v\n", acct, err), "", "BankAdapter - GetBankAccountByAccountNumber")
		log.Error().Msg(logErr)
		return bankAccountOrm, err
	}

	return bankAccountOrm, nil
}

func (a *DatabaseAdapter) InsertExchangeRate(r domainBank.BankExchangeRateOrm) (uuid.UUID, error) {
	if err := a.db.Create(&r).Error; err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't insert exchange rate : %v\n", err), "", "BankAdapter - InsertExchangeRate")
		log.Error().Msg(logErr)
		return uuid.Nil, err
	}

	// log success
	log.Info().Msgf("Exchange rate inserted with uuid %v", r.ExchangeRateUuid)

	return r.ExchangeRateUuid, nil
}

func (a *DatabaseAdapter) GetExchangeRateAtTimestamp(fromCurrency string, toCurrency string, ts time.Time) (domainBank.BankExchangeRateOrm, error) {
	var exchangeRateOrm domainBank.BankExchangeRateOrm

	err := a.db.First(&exchangeRateOrm, "from_currency = ? AND to_currency = ? AND (? BETWEEN valid_from_timestamp and valid_to_timestamp)", fromCurrency, toCurrency, ts).Error

	if err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't find exchange rate from %v to %v at %v : %v\n", fromCurrency, toCurrency, ts, err), "", "BankAdapter - GetExchangeRateAtTimestamp")
		log.Error().Msg(logErr)
	}

	return exchangeRateOrm, err
}

func (a *DatabaseAdapter) CreateTransaction(account domainBank.BankAccountOrm, trx domainBank.BankTransactionOrm) (uuid.UUID, error) {
	tx := a.db.Begin()

	newAmount := trx.Amount

	if trx.TransactionType == domainBank.TransactionTypeOut {
		newAmount = -1 * newAmount
	}

	// Create the transaction
	if err := tx.Create(trx).Error; err != nil {
		tx.Rollback()
		logErr := util.LogError(fmt.Sprintf("Can't create transaction : %v\n", err), "", "BankAdapter - CreateTransaction")
		log.Error().Msg(logErr)
		return uuid.Nil, err
	}

	newAccountBalance := account.CurrentBalance + newAmount

	// update account balance
	if err := tx.Model(&account).Updates(
		map[string]interface{}{
			"current_balance": newAccountBalance,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	tx.Commit()

	return trx.TransactionUuid, nil
}

func (a *DatabaseAdapter) CreateTransfer(trf domainBank.BankTransferOrm) (uuid.UUID, error) {
	if err := a.db.Create(trf).Error; err != nil {
		return uuid.Nil, err
	}

	return trf.TransferUuid, nil
}

func (a *DatabaseAdapter) CreateTransferTransactionPair(fromAccountOrm domainBank.BankAccountOrm, toAccountOrm domainBank.BankAccountOrm,
	fromTransactionOrm domainBank.BankTransactionOrm, toTransactionOrm domainBank.BankTransactionOrm) (bool, error) {
	tx := a.db.Begin()

	// from account
	if err := tx.Create(fromTransactionOrm).Error; err != nil {
		return false, err
	}

	// to account
	if err := tx.Create(toTransactionOrm).Error; err != nil {
		return false, err
	}

	// recalculate balance from account
	newBalanceFrom := fromAccountOrm.CurrentBalance - fromTransactionOrm.Amount

	// recalculate balance to account
	newBalanceTo := toAccountOrm.CurrentBalance + toTransactionOrm.Amount

	if err := tx.Model(&fromAccountOrm).Updates(
		map[string]interface{}{
			"current_balance": newBalanceFrom,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Model(&toAccountOrm).Updates(
		map[string]interface{}{
			"current_balance": newBalanceTo,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	tx.Commit()

	return true, nil
}

func (a *DatabaseAdapter) UpdateTransferStatus(transfer domainBank.BankTransferOrm, status bool) error {
	if err := a.db.Model(&transfer).Updates(
		map[string]interface{}{
			"transfer_success": status,
			"updated_at":       time.Now(),
		},
	).Error; err != nil {
		return err
	}

	return nil
}
