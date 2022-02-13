package database

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database holds information for the database connection.
type Database struct {
	gormdb             *gorm.DB
	sqldb              *sql.DB
	credentialsManager *credentials.Manager
}

func createFileDir(file string) {
	dir := filepath.Dir(file)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// nosemgrep: tests.semgrep.go.lang.correctness.permissions.incorrect-default-permission
		if err := os.MkdirAll(dir, 0750); err != nil {
			panic(err)
		}
	}
}

// Create instanciates a database connection.
func Create(cm *credentials.Manager, dialect, connection string) (*Database, error) {
	log.Println("Setting up database connection.")

	maxOpenConns := 5

	var db *gorm.DB
	var err error

	switch dialect {
	case "sqlite3":
		createFileDir(connection)
		maxOpenConns = 1
		db, err = gorm.Open(sqlite.Open(connection), &gorm.Config{})
	case "mysql":
		if !strings.Contains(connection, "parseTime=true") {
			connection += "?parseTime=true"
		}
		db, err = gorm.Open(mysql.Open(connection), &gorm.Config{})
	default:
		message := "Database dialect is not supported"
		return nil, errors.New(message)
	}

	if err != nil {
		return nil, err
	}

	sql, err := db.DB()
	if err != nil {
		return nil, err
	}

	sql.SetMaxOpenConns(maxOpenConns)

	if dialect == "mysql" {
		sql.SetConnMaxLifetime(9 * time.Minute)
	}

	err = db.AutoMigrate(&model.User{}, &model.Application{})
	if err != nil {
		return nil, err
	}

	return &Database{gormdb: db, sqldb: sql, credentialsManager: cm}, nil
}

// Close closes the database connection.
func (d *Database) Close() {
	err := d.sqldb.Close()
	if err != nil {
		log.Printf("Error while closing database: %s", err)
	}
}

// Populate fills the database with initial information like the admin user.
func (d *Database) Populate(name, password, matrixID string) error {
	log.Print("Populating database.")

	var user model.User

	query := d.gormdb.Where("name = ?", name).First(&user)

	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		user, err := model.NewUser(d.credentialsManager, name, password, true, matrixID)
		if err != nil {
			return err
		}

		if err := d.gormdb.Create(&user).Error; err != nil {
			return errors.New("user cannot be created")
		}
	} else {
		log.Printf("Priviledged user %s already exists.", name)
	}

	return nil
}

// RepairChannels resets channels that have been modified by a user.
func (d *Database) RepairChannels(dp Dispatcher) error {
	log.Print("Repairing application channels.")

	users, err := d.GetUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		user := user // See https://stackoverflow.com/a/68247837

		applications, err := d.GetApplications(&user)
		if err != nil {
			return err
		}

		for _, application := range applications {
			application := application // See https://stackoverflow.com/a/68247837

			if err := dp.UpdateApplication(&application); err != nil {
				return err
			}

			orphan, err := dp.IsOrphan(&application, &user)
			if err != nil {
				return err
			}

			if orphan {
				log.Printf("Found orphan channel for application %s (ID %d)", application.Name, application.ID)

				if err = dp.RepairApplication(&application, &user); err != nil {
					log.Printf("Unable to repair application %s (ID %d).", application.Name, application.ID)
					log.Println(err)
				}
			}
		}
	}

	return nil
}

// GetSqldb returns the databases sql.DB object
func (d *Database) GetSqldb() *sql.DB {
	return d.sqldb
}
