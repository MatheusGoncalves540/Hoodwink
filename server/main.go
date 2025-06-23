package main

import (
	"log"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/config"
	"github.com/MatheusGoncalves540/Hoodwink/db"
	"github.com/MatheusGoncalves540/Hoodwink/router"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Erro ao carregar .env")
	}
	config.CheckEnvVars(".env.example")
	routes := router.SetupRoutes()

	db.ConncetDB()

	log.Println("Servidor ouvindo em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", routes))
}
