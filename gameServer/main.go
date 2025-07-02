package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/rHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Erro ao carregar .env")
	}

	redisClient := redis.ConnectRedis()

	services := services.SetupServices(redisClient)
	handler := rHandlers.NewHandler(services)

	config.CheckEnvVars(".env.example")
	routes := routes.SetupRoutes(handler)

	log.Printf("Servidor ouvindo em %s", os.Getenv("GAME_SERVER_URL"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), routes))
}
