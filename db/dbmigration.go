package dbmigration

import (
	"database/sql"

	"github.com/fajaramaulana/go-grpc-micro-bank-server/util"
	migrate "github.com/golang-migrate/migrate/v4"

	"github.com/chilts/sid"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

func Migrate(conn *sql.DB) {
	sidString := sid.Id()
	logStart := util.LogRequest("", "Migrate-"+sidString, "dbmigration - Migrate")
	log.Info().Msg(logStart)

	dbDriver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		logErr := util.LogError(err.Error(), "Migrate-"+sidString, "dbmigration - Migrate - postgres.WithInstance")
		log.Fatal().Msg(logErr)
	}

	m, err := migrate.NewWithDatabaseInstance("file://../db/migrations", "postgres", dbDriver)

	if err != nil {
		logErr := util.LogError(err.Error(), "Migrate-"+sidString, "dbmigration - Migrate - migrate.NewWithDatabaseInstance")
		log.Fatal().Msg(logErr)
	}

	// if err := m.Down(); err != nil {
	// 	log.Println("Database migration (down) failed :", err)
	// }

	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			log.Info().Msg("Database migration (up) no change")
		} else {
			logErr := util.LogError(err.Error(), "Migrate-"+sidString, "dbmigration - Migrate - m.Up()")
			log.Fatal().Msg(logErr)
		}
	}

	logStop := util.LogResponse("Database Migration Done", "Migrate-"+sidString, "dbmigration - Migrate")
	log.Info().Msg(logStop)
}
