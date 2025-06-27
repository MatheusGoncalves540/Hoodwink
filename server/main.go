package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	db.ConnectDB()

	log.Printf("Servidor ouvindo em %s", os.Getenv("SERVER_URL"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), routes))
}
