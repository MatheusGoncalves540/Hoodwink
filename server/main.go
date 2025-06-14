package main

import (
	"log"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	utils.CheckEnvVars(".env.example")
	routes := router.SetupRoutes()

	log.Println("Servidor ouvindo em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", routes))
}
