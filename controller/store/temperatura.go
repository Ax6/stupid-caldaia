package store

import (
	"context"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PrimaryRetentionTime = 7 * 24 * 60 * 60 * 1000 // 7 days
	CompactTime          = 5 * 60 * 1000           // 5 minutes
)

type SensorOptions struct {
	Name     string
	Position string
}

type Sensor struct {
	Name         string
	Position     string
	Client       *redis.Client
	key          string
	compactedKey string
}

type Measure struct {
	Timestamp float64
	Value     float64
}

func NewSensor(ctx context.Context, client *redis.Client, opt *SensorOptions) (*Sensor, error) {
	key := opt.Name + ":" + opt.Position
	compactedKey := opt.Name + "_compacted" + ":" + opt.Position
	sensor := Sensor{opt.Name, opt.Position, client, key, compactedKey}

	// Check if sensor already exists
	exists, _ := sensor.Client.Exists(ctx, key).Result()
	if exists == 0 {
		// If not, create it
		_, err := sensor.Client.TSCreateWithArgs(ctx, key, &redis.TSOptions{
			Retention: PrimaryRetentionTime,
			Labels:    map[string]string{"position": sensor.Position},
		}).Result()
		if err != nil {
			return &sensor, err
		}
	}
	exists, _ = sensor.Client.Exists(ctx, compactedKey).Result()
	if exists == 0 {
		// Create compaction key
		_, err := sensor.Client.TSCreateWithArgs(ctx, compactedKey, &redis.TSOptions{Retention: 0}).Result()
		if err != nil {
			return &sensor, err
		}

		// Create compaction rule
		_, err = sensor.Client.TSCreateRule(ctx, key, compactedKey, redis.Avg, CompactTime).Result()
		if err != nil {
			return &sensor, err
		}
	}

	return &sensor, nil
}

func (s *Sensor) Add(ctx context.Context, measure *Measure) error {
	_, err := s.Client.TSAdd(ctx, s.key, measure.Value, measure.Timestamp).Result()
	if err != nil {
		return err
	}
	return nil
}

func (s *Sensor) Sample(ctx context.Context) (*Measure, error) {
	// Create a random number for testing
	value := rand.Float64() * 100
	time := float64(time.Now().UnixMilli())
	sample := Measure{Timestamp: time, Value: value}
	s.Client.Publish(ctx, s.key, value)
	return &sample, nil
}
