package main

import (
	"log"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router"
)

func main() {
	r := router.SetupRoutes()

	log.Println("Servidor ouvindo em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
