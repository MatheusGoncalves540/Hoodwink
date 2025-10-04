package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/routeHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Erro ao carregar .env")
	}
	config.CheckEnvVars(".env.example")

	redisClient := redis.ConnectRedis()

	services := services.SetupServices(redisClient)
	handler := routeHandlers.NewHandler(services)

	// Inicia o serviço de heartbeat
	services.HeartbeatService.Start()
	defer services.HeartbeatService.Stop()

	routes := routes.SetupRoutes(handler, redisClient)

	log.Printf("Servidor ouvindo em %s (Instância: %s)", os.Getenv("GAME_SERVER_URL"), config.InstanceID)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), routes))
}
