package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	dbHost := os.Getenv("PG_HOST")
	dbName := os.Getenv("PG_NAME")
	dbUser := os.Getenv("PG_USER")
	dbPass := os.Getenv("PG_PASS")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}

	// Pool: teto de conexões abertas/ociosas e reciclagem periódica.
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal("Erro ao obter pool do banco:", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if os.Getenv("AUTOMIGRATE") == "true" {
		database.AutoMigrate(
			&models.User{},
			&models.ApiKey{},
		)
	}

	return database, nil
}
