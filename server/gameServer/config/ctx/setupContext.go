package ctx

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/http"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
)

func SetupContext() *RoutesContext {
	redisClient := redis.ConnectRedis()

	services := services.SetupServices(redisClient)
	handler := http.NewHandler(services)
	rulesRegistry := rules.SetupRulesRegistry("game/rules/modes")
	routesContext := &RoutesContext{
		Rdb:           redisClient,
		Services:      services,
		Handler:       handler,
		RulesRegistry: rulesRegistry,
	}
	return routesContext
}
