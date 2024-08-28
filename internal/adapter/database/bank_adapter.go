package database

import (
	"fmt"

	"github.com/fajaramaulana/go-grpc-micro-bank-server/model/orm"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/rs/zerolog/log"
)

func (a *DatabaseAdapter) GetBalanceBankAccountByAccountNumber(acct string) (orm.BalanceAccountOrm, error) {
	var bankAccountOrm orm.BalanceAccountOrm

	if err := a.db.Select("account_uuid, account_number, currency, current_balance").First(&bankAccountOrm, "account_number = ?", acct).Error; err != nil {
		logErr := util.LogError(fmt.Sprintf("Can't find bank account number %v : %v\n", acct, err), "", "DatabaseAdapter - GetBankAccountByAccountNumber")
		log.Error().Msg(logErr)
		return bankAccountOrm, err
	}

	return bankAccountOrm, nil
}
