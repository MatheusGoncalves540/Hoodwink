package ctx

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/http"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/redis/go-redis/v9"
)

type RoutesContext struct {
	Rdb           *redis.Client
	Services      *services.Services
	Handler       *http.HTTPHandler
	RulesRegistry *rules.Registry
}
