package database

import (
	"fmt"
	"time"

	domainBank "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application/domain/bank"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (a *DatabaseAdapter) GetBalanceBankAccountByAccountNumber(acct string) (domainBank.BalanceAccountOrm, error) {
	var bankAccountOrm domainBank.BalanceAccountOrm

	if err := a.db.Select("account_uuid, account_number, currency, current_balance").First(&bankAccountOrm, "account_number = ?", acct).Error; err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't find bank account number %v : %v\n", acct, err), "", "DatabaseAdapter - GetBankAccountByAccountNumber")
		log.Error().Msg(logErr)
		return bankAccountOrm, err
	}

	return bankAccountOrm, nil
}

func (a *DatabaseAdapter) InsertExchangeRate(r domainBank.BankExchangeRateOrm) (uuid.UUID, error) {
	if err := a.db.Create(&r).Error; err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't insert exchange rate : %v\n", err), "", "DatabaseAdapter - InsertExchangeRate")
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
		logErr := util.LogError(fmt.Sprintf("Can't find exchange rate from %v to %v at %v : %v\n", fromCurrency, toCurrency, ts, err), "", "DatabaseAdapter - GetExchangeRateAtTimestamp")
		log.Error().Msg(logErr)
	}

	return exchangeRateOrm, err

}
