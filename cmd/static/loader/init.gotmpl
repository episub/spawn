package loader

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

var (
	// ErrInternal Safe message for end users indicating an internal issue
	ErrInternal = fmt.Errorf("Internal error, please try again or contact us.")

	// ErrUnsupportedField Field not supported
	ErrUnsupportedField = fmt.Errorf("Field is not enabled for this action, please contact support if this is not correct")

	// ErrNoRecord Returned when the result could not be found
	ErrNoRecords = fmt.Errorf("No such record(s) could be found")
)

// PostgresLoader Loader using postgres database
type PostgresLoader struct {
	pool   *pgx.ConnPool
	config pgx.ConnConfig
}

// Loader Stores the currently configured loader.  You could optionally create an interface for loaders and swap them between each other
var Loader PostgresLoader

// InitialiseLoader Set up loader with the correct database values
func InitialiseLoader(dbName string, dbUser string, dbPass string, dbHost string, givenLog *logrus.Logger) error {
	log = givenLog
	connString := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable", dbUser, dbName, dbPass, dbHost)
	connConfig, err := pgx.ParseConnectionString(connString)

	if err != nil {
		return err
	}

	Loader.config = connConfig

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: connConfig, MaxConnections: 5 /* https://wiki.postgresql.org/wiki/Number_Of_Database_Connections#How_to_Find_the_Optimal_Database_Connection_Pool_Size */})

	if err != nil {
		return err
	}

	Loader.pool = pool

	runBatchLoaders()

	return nil
}

func sanitiseError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case err == pgx.ErrNoRows || err == sql.ErrNoRows:
		return ErrNoRecords
	case strings.Contains(err.Error(), "invalid input syntax for type"):
		log.Errorf("Sanitised error: %s", err)
		return fmt.Errorf("One or more provided values are invalid.  Please check inputs.")
	case strings.Contains(err.Error(), "Expected 1 row, but had 0"):
		return fmt.Errorf("Could not find item")
	case err.Error() == "Email address already used":
		return err
	case strings.Contains(err.Error(), "invalid input syntax for uuid"):
		log.Printf("Error: %s", err)
		return fmt.Errorf("One or more fields have missing or invalid required id")
	default:
		log.WithField("error", err).Error("Unhandled error")
		return fmt.Errorf("Unknown error.  Please contact support.")
	}
}
