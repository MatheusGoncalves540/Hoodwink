package ctx

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/http"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
)

func SetupContext() *RoutesContext {
	redisClient := redisFuncs.ConnectRedis()

	services := services.SetupServices(redisClient)
	rulesRegistry := rules.SetupRulesRegistry("game/rules/modes")
	handler := http.NewHandler(services, rulesRegistry)
	routesContext := &RoutesContext{
		Rdb:           redisClient,
		Services:      services,
		Handler:       handler,
		RulesRegistry: rulesRegistry,
	}
	return routesContext
}
