package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config/ctx"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	if os.Getenv("ENVIRONMENT") == "local" {
		utils.LogDebug("⚠️ Ambiente definido como local")
		config.CheckEnvVars(".env.example")
	}

	routesContext := ctx.SetupContext()
	routes := routes.SetupRoutes(routesContext)

	engine.StartGameProcessor(routesContext.Rdb, routesContext.RulesRegistry)

	// Inicia o serviço de heartbeat
	routesContext.Services.HeartbeatService.Start()
	defer routesContext.Services.HeartbeatService.Stop()

	log.Printf("Servidor ouvindo em %s (Instância: %s)", os.Getenv("GAME_SERVER_URL"), config.InstanceID)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), routes))
}
