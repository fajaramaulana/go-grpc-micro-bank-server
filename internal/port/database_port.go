package port

import (
	"github.com/fajaramaulana/go-grpc-micro-bank-server/model/orm"
)

type BankDatabasePort interface {
	GetBalanceBankAccountByAccountNumber(acct string) (orm.BalanceAccountOrm, error)
}
