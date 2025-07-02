package redis

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
)

func ConnectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),     // localhost:6379
		Password: os.Getenv("REDIS_PASSWORD"), // vazio se não tiver
		DB:       0,                           // database padrão
	})

	// Testa a conexão
	_, err := client.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Erro ao conectar ao Redis: %v", err)
	}

	log.Println("Conectado ao Redis com sucesso!")
	return client
}
