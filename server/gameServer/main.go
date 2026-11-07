package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config/ctx"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/runtime"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func main() {
	config.SetupEnvs()

	routesContext := ctx.SetupContext()
	router := routes.SetupRoutes(routesContext)

	runtime.StartGameProcessor(routesContext.Rdb, routesContext.RulesRegistry)

	// Inicia o serviço de heartbeat
	routesContext.Services.HeartbeatService.Start()
	defer routesContext.Services.HeartbeatService.Stop()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: router,
	}

	go func() {
		log.Printf("📡 Servidor ouvindo em %s (Instância: %s)", os.Getenv("GAME_SERVER_URL"), utils.GetInstanceID())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servidor: %v", err)
		}
	}()

	<-sigCtx.Done()
	log.Println("Encerrando: drenando request logs...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	routesContext.Services.RequestLogShipper.Close()
}
