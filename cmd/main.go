// Package main is the entry point of the application.
// It initializes the necessary components and starts the gRPC server.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	cfg "github.com/fajaramaulana/go-grpc-micro-bank-server/config"
	dbmigration "github.com/fajaramaulana/go-grpc-micro-bank-server/db"
	mydb "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/adapter/database"
	mygrpc "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/adapter/grpc"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application"
	domainBank "github.com/fajaramaulana/go-grpc-micro-bank-server/internal/application/domain/bank"
	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/rand"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/chilts/sid"
)

// main is the entry point of the application.
// It initializes the necessary components and starts the gRPC server.
func main() {
	sidString := sid.Id()
	// Configure the logger to output logs to the console
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load the configuration from the .env file
	configuration := cfg.New("../.env")

	// Create the connection string for the database
	conn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s", configuration.Get("DB_DRIVER"), configuration.Get("DB_USER"), configuration.Get("DB_PASSWORD"), configuration.Get("DB_HOST"), configuration.Get("DB_PORT"), configuration.Get("DB_NAME"), configuration.Get("DB_SSLMODE"))

	// Open a connection to the database
	sqlDb, err := sql.Open("pgx", conn)
	if err != nil {
		logErr := util.LogError(err.Error(), "Main-"+sidString, "Main - sql.Open")
		log.Fatal().Msg(logErr)
	}

	// Run database migrations
	dbmigration.Migrate(sqlDb)

	databaseAdapter, err := mydb.NewDatabaseAdapter(sqlDb)

	if err != nil {
		logErr := util.LogError(err.Error(), "Main-"+sidString, "Main - mydb.NewDatabaseAdapter")
		log.Fatal().Msg(logErr)
	}

	// Create an instance of the BankService
	bankService := application.NewBankService(databaseAdapter)

	// Create an instance of the BankService
	resilliencyService := application.NewResilliencyService()

	go generateExchangeRates(bankService, "USD", "IDR", 5*time.Second)
	// Create a gRPC adapter with the BankService and start the server

	portInt, err := strconv.Atoi(configuration.Get("PORT"))
	if err != nil {
		logErr := util.LogError(err.Error(), "Main-"+sidString, "Main - Conv String to int Port")
		log.Fatal().Msg(logErr)
	}
	grpcAdapter := mygrpc.NewGrpcAdapter(bankService, resilliencyService, portInt)
	grpcAdapter.Run()
}

func generateExchangeRates(bs *application.BankService, fromCurrency, toCurrency string, duration time.Duration) {
	ticker := time.NewTicker(duration)

	for range ticker.C {
		now := time.Now()
		validFrom := now.Truncate(time.Second).Add(3 * time.Second)
		validTo := validFrom.Add(duration).Add(-1 * time.Millisecond)

		dummyRate := domainBank.ExchangeRate{
			FromCurrency:       fromCurrency,
			ToCurrency:         toCurrency,
			ValidFromTimestamp: validFrom,
			ValidToTimestamp:   validTo,
			Rate:               2000 + float64(rand.Intn(300)),
		}

		bs.CreateExchangeRate(dummyRate)

		dummyRateReverse := domainBank.ExchangeRate{
			FromCurrency:       toCurrency,
			ToCurrency:         fromCurrency,
			ValidFromTimestamp: validFrom,
			ValidToTimestamp:   validTo,
			Rate:               1 / dummyRate.Rate,
		}

		bs.CreateExchangeRate(dummyRateReverse)
	}
}
