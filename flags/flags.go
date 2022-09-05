package flags

import (
	"flag"
	"fmt"
	"golang-migrate-cli/constants"
	"log"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

var validDrivers = []string{constants.SQLITE3, constants.POSTGRES, constants.MYSQL}
var validOperations = []string{constants.UP, constants.DOWN}
var Operation = flag.String(
	"operation",
	constants.UP,
	fmt.Sprintf("A migration operation to execute with the environment database. Possible values: %s. Default: %s", strings.Join(validOperations, ", "), constants.UP),
)
var ExecutionCount = flag.Int("execution-count", 0, "Count of migrations that must execute. Default = all")
var MustDrop = flag.Bool("drop", false, "Drops the database (USE WITH CAUTION). Default = false")
var MigrationsPath = flag.String("path", "", "Migrations folder path. If not specified, the \"DATABASE_MIGRATIONS_FOLDER_PATH\" environment variable is used")
var DatabaseDriver = flag.String("db-driver", constants.POSTGRES, fmt.Sprintf("Database driver. Possible values: %s. Default: %s", strings.Join(validDrivers, ", "), constants.POSTGRES))
var DatabaseURI = flag.String("db-uri", "", "Database connection URI.")
var EnvfilePath = flag.String("env-file", "", "Environment variables file.")
var SSLMode = flag.String("ssl-mode", "disable", "Database SSL connection mode. Default: disable")

func init() {
	flag.Parse()
	validateFlags()
}

func validateFlags() {
	if *EnvfilePath != "" {
		godotenv.Load(*EnvfilePath)
	}
	validateOperationFlag()
	validateDatabaseURIFlag()
}

func validateOperationFlag() {
	var operationIsValid bool
	for _, validOperation := range validOperations {
		operationIsValid = *Operation == validOperation
		if operationIsValid {
			break
		}
	}
	if !operationIsValid {
		log.Fatalf("Invalid operation. The available operations are: %s.", strings.Join(validOperations, ", "))
	}
}

func validateDatabaseURIFlag() {
	if *DatabaseURI == "" {
		return
	} else if *DatabaseDriver == constants.SQLITE3 {
		log.Fatalf("When using the \"%s\" driver, you must provide only the database host!", constants.SQLITE3)
	}
	uriCompile := regexp.MustCompile("^(.+)://")
	uriMatch := uriCompile.FindStringSubmatch(*DatabaseURI)
	if uriMatch == nil {
		log.Fatal("You must provide a valid database connection URI!")
	}
	var driverIsValid bool
	for _, driver := range validDrivers {
		if driver == uriMatch[1] {
			driverIsValid = true
			*DatabaseDriver = uriMatch[1]
			break
		}
	}
	if !driverIsValid {
		log.Fatal("You must provide a valid database driver!")
	}
}
