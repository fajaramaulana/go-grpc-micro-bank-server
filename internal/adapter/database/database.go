package database

import (
	"database/sql"
	"fmt"

	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseAdapter struct {
	db *gorm.DB
}

func NewDatabaseAdapter(conn *sql.DB) (*DatabaseAdapter, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: conn,
	}), &gorm.Config{})

	if err != nil {
		logErr := util.LogError(fmt.Sprintf("can't connect database (gorm) : %v", err), "", "DatabaseAdapter - GetBankAccountByAccountNumber")
		log.Error().Msg(logErr)
		return nil, fmt.Errorf("can't connect database (gorm) : %v", err)
	}

	return &DatabaseAdapter{
		db: db,
	}, nil
}
