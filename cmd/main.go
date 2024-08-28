package main

import (
	"database/sql"
	"fmt"
	"os"

	cfg "github.com/fajaramaulana/go-grpc-micro-bank-server/config"
	dbmigration "github.com/fajaramaulana/go-grpc-micro-bank-server/db"
	mygrpc "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/adapter/grpc"
	app "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	configuration := cfg.New("../.env")

	conn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s", configuration.Get("DB_DRIVER"), configuration.Get("DB_USER"), configuration.Get("DB_PASSWORD"), configuration.Get("DB_HOST"), configuration.Get("DB_PORT"), configuration.Get("DB_NAME"), configuration.Get("DB_SSLMODE"))
	sqlDb, err := sql.Open("pgx", conn)
	if err != nil {
		log.Fatal().Msgf("Error opening database: %v", err)
	}

	dbmigration.Migrate(sqlDb)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	bankService := &app.BankService{}

	grpcAdapter := mygrpc.NewGrpcAdapter(bankService, 8080)
	grpcAdapter.Run()
}
