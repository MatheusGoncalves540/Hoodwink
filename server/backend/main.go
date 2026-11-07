package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/config"
	"github.com/MatheusGoncalves540/Hoodwink/db"
	"github.com/MatheusGoncalves540/Hoodwink/routes"
	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/joho/godotenv"
)

func main() {
	runRetentionOnce := flag.Bool("run-retention-once", false, "roda a retenção de request logs uma vez e sai")
	flag.Parse()

	godotenv.Load(".env")
	log.Printf("🌐 Ambiente definido como %s", os.Getenv("ENVIRONMENT"))

	if os.Getenv("ENVIRONMENT") == "local" {
		config.CheckEnvVars(".env.example")
	}

	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	svc := services.SetupServices(database)
	handler := handlers.NewHandler(svc)
	router := routes.SetupRoutes(handler)

	if *runRetentionOnce {
		if err := services.RunRetention(context.Background(), database); err != nil {
			log.Fatalf("retenção: %v", err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: router,
	}

	go func() {
		log.Printf("Servidor ouvindo em %s", os.Getenv("SERVER_URL"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servidor: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Encerrando: drenando request logs...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}

	// Drena o canal e grava o último lote (timeout de 3s por INSERT).
	svc.RequestLogService.Close()
}
