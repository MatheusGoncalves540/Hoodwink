package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink/config"
	"github.com/MatheusGoncalves540/Hoodwink/db"
	"github.com/MatheusGoncalves540/Hoodwink/routes"
	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	if os.Getenv("ENVIRONMENT") == "local" {
		utils.LogDebug("⚠️ Ambiente definido como local")
		config.CheckEnvVars(".env.example")
	}

	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	services := services.SetupServices(database)
	handler := handlers.NewHandler(services)
	routes := routes.SetupRoutes(handler)

	log.Printf("Servidor ouvindo em %s", os.Getenv("SERVER_URL"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), routes))
}
