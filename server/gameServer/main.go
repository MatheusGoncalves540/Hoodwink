package gameserver

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Erro ao carregar .env")
	}
	database, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	services := services.SetupServices(database)
	handler := handlers.NewHandler(services)

	config.CheckEnvVars(".env.example")
	routes := routes.SetupRoutes(handler)

	log.Printf("Servidor ouvindo em %s", os.Getenv("SERVER_URL"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), routes))
}
