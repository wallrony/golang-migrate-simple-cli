package main

import (
	"database/sql"
	"fmt"
	"golang-migrate-cli/constants"
	"golang-migrate-cli/flags"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Config struct {
	MigrationsTable string
	DatabaseName    string
}

func init() {
	validateMigrationsPath()
}

func validateMigrationsPath() {
	if *flags.MigrationsPath == "" && os.Getenv("DATABASE_MIGRATIONS_FOLDER_PATH") == "" {
		log.Fatal("You must specify the migrations folder path using the \"--path\" flag or setting the \"DATABASE_MIGRATIONS_FOLDER_PATH\" environment variable.")
	}
}

func main() {
	if *flags.MustDrop {
		dropDatabase()
	} else if *flags.Operation == constants.DOWN {
		downMigrations()
	} else {
		upMigrations()
	}
}

func dropDatabase() {
	fmt.Println("Dropping database...")
	migrationManager, db := getMigrationManager()
	defer db.Close()
	var err error
	err = migrationManager.Drop()
	if err != nil {
		dealWithMigrationError(err)
	}
}

func downMigrations() {
	fmt.Println("Executing down migrations...")
	migrationManager, db := getMigrationManager()
	defer db.Close()
	var err error
	if *flags.ExecutionCount == 0 {
		err = migrationManager.Down()
	} else {
		err = migrationManager.Steps(*flags.ExecutionCount * -1)
	}
	if err != nil {
		dealWithMigrationError(err)
	}
}

func upMigrations() {
	fmt.Println("Executing up migrations...")
	migrationManager, db := getMigrationManager()
	defer db.Close()
	var err error
	if *flags.ExecutionCount == 0 {
		err = migrationManager.Up()
	} else {
		err = migrationManager.Steps(*flags.ExecutionCount)
	}
	if err != nil {
		dealWithMigrationError(err)
	}
}

func getDatabaseURI() string {
	if *flags.DatabaseURI != "" {
		return *flags.DatabaseURI
	}
	host := os.Getenv("DATABASE_HOST")
	if *flags.DatabaseDriver == constants.SQLITE3 {
		return host
	}
	database := os.Getenv("DATABASE_NAME")
	port := os.Getenv("DATABASE_PORT")
	username := os.Getenv("DATABASE_USERNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	authentication := fmt.Sprintf("%s:%s@", username, password)
	return fmt.Sprintf("%s://%s%s:%s/%s?sslmode=%s", *flags.DatabaseDriver, authentication, host, port, database, *flags.SSLMode)
}

func getMigrationManager() (*migrate.Migrate, *sql.DB) {
	uri := getDatabaseURI()
	db, err := sql.Open(*flags.DatabaseDriver, uri)
	if err != nil {
		log.Fatalf("An error occurred when connecting to the database: %s", err.Error())
	}
	var migrationManager *migrate.Migrate
	var migrationsPath = getMigrationsPath()
	var driver = getDatabaseDriver(db)
	var databaseName = os.Getenv("DATABASE_NAME")
	if *flags.DatabaseDriver == constants.SQLITE3 {
		fileSource, err := (&file.File{}).Open(migrationsPath)
		if err != nil {
			log.Fatal(err)
		}
		migrationManager, err = migrate.NewWithInstance(
			"file",
			fileSource,
			databaseName,
			driver,
		)
	} else {
		migrationManager, err = migrate.NewWithDatabaseInstance(
			fmt.Sprintf("file://%s/%s", migrationsPath, *flags.Operation),
			databaseName,
			driver,
		)
	}
	if err != nil {
		log.Fatalf("An error occurred when creating the migration manager: %s", err.Error())
	}
	return migrationManager, db
}

func getDatabaseDriver(db *sql.DB) database.Driver {
	var driver database.Driver
	var err error
	if *flags.DatabaseDriver == constants.SQLITE3 {
		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	} else if *flags.DatabaseDriver == constants.MYSQL {
		driver, err = mysql.WithInstance(db, &mysql.Config{})
	} else {
		driver, err = postgres.WithInstance(db, &postgres.Config{})
	}
	if err != nil {
		log.Fatalf("An error occurred when getting the database connection driver: %s", err.Error())
	}
	return driver
}

func getMigrationsPath() string {
	if *flags.MigrationsPath != "" {
		return *flags.MigrationsPath
	}
	return os.Getenv("DATABASE_MIGRATIONS_FOLDER_PATH")
}

func dealWithMigrationError(err error) {
	if !strings.Contains(err.Error(), "no change") {
		if !strings.Contains(err.Error(), "no change") {
			log.Fatalf("An error occurred when migrating: %s", err.Error())
		}
		fmt.Println("No new migrations were found.")
	}
}
