package testutils

import (
	"context"
	"fmt"
	"stupid-caldaia/controller/graph/model"
	"testing"

	"github.com/redis/go-redis/v9"
)

const (
	MIN_TEMP    = 10
	MAX_TEMP    = 20
	LOCAL_REDIS = "localhost:6379"
)

func CreateTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: LOCAL_REDIS})
}

func CreateTestBoiler(ctx context.Context, t *testing.T) (*model.Boiler, error) {
	// Make sure we clean up before creating a new boiler
	client := CreateTestRedis()
	err := client.Del(
		ctx,
		"test_boiler_"+t.Name(),
		"switch:"+"test_boiler_"+t.Name(),
		"overheating:"+"test_boiler_"+t.Name(),
	).Err()
	if err != nil {
		return nil, err
	}

	boiler, err := model.NewBoiler(ctx, client, model.BoilerConfig{
		Name:                  fmt.Sprintf("test_boiler_%s", t.Name()),
		DefaultMinTemperature: MIN_TEMP,
		DefaultMaxTemperature: MAX_TEMP,
		SwitchPin:             1,
	})
	if err != nil {
		return nil, err
	}
	return boiler, nil
}
