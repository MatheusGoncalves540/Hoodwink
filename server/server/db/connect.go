package db

import (
	"fmt"
	"log"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConncetDB() {
	dbHost := os.Getenv("PG_HOST")
	dbName := os.Getenv("PG_NAME")
	dbUser := os.Getenv("PG_USER")
	dbPass := os.Getenv("PG_PASS")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}

	database.AutoMigrate(&models.User{})

	DB = database
}
