package graph

import (
	"stupid-caldaia/controller/graph/model"

	"github.com/redis/go-redis/v9"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Boiler  *model.Boiler
	Client  *redis.Client
	Sensors map[string]*model.Sensor
}
