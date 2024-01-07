package internal

import (
	"context"
	"fmt"
	"math/rand"
	"stupid-caldaia/controller/graph/model"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

const (
	MIN_TEMP = 10
	MAX_TEMP = 20
)

func CreateTestBoiler(t *testing.T, ctx context.Context) (*model.Boiler, error) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	boiler, err := model.NewBoiler(ctx, client, model.BoilerConfig{
		Name:                  fmt.Sprintf("test_boiler_%d", rand.Int()),
		DefaultMinTemperature: MIN_TEMP,
		DefaultMaxTemperature: MAX_TEMP,
		SwitchPin:             1,
	})
	if err != nil {
		return nil, err
	}
	return boiler, nil
}
