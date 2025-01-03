package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.56

import (
	"context"
	"slices"
	"stupid-caldaia/controller/graph/model"
	"time"
)

// UpdateBoiler is the resolver for the updateBoiler field.
func (r *mutationResolver) UpdateBoiler(ctx context.Context, state *model.State, minTemp *float64, maxTemp *float64) (*model.BoilerInfo, error) {
	if state != nil {
		_, err := r.Resolver.Boiler.Switch(ctx, *state)
		if err != nil {
			return nil, err
		}
	}

	if minTemp != nil {
		_, err := r.Resolver.Boiler.SetMinTemp(ctx, *minTemp)
		if err != nil {
			return nil, err
		}
	}

	if maxTemp != nil {
		_, err := r.Resolver.Boiler.SetMaxTemp(ctx, *maxTemp)
		if err != nil {
			return nil, err
		}
	}

	return r.Resolver.Boiler.GetInfo(ctx)
}

// SetRule is the resolver for the setRule field.
func (r *mutationResolver) SetRule(ctx context.Context, id *string, start time.Time, duration time.Duration, delay time.Duration, targetTemp float64, repeatDays []int) (*model.Rule, error) {
	slices.Sort(repeatDays)
	opt := &model.Rule{
		Start:      start,
		Duration:   duration,
		Delay:      delay,
		TargetTemp: targetTemp,
		RepeatDays: repeatDays,
	}
	if id != nil {
		opt.ID = *id
	}
	return r.Resolver.Boiler.SetRule(ctx, opt)
}

// StopRule is the resolver for the stopRule field.
func (r *mutationResolver) StopRule(ctx context.Context, id string) (bool, error) {
	programmedInterval, err := r.Resolver.Boiler.StopRule(ctx, id)
	return !programmedInterval.IsActive, err
}

// DeleteRule is the resolver for the deleteRule field.
func (r *mutationResolver) DeleteRule(ctx context.Context, id string) (bool, error) {
	err := r.Resolver.Boiler.DeleteRule(ctx, id)
	return err == nil, err
}

// Boiler is the resolver for the boiler field.
func (r *queryResolver) Boiler(ctx context.Context) (*model.BoilerInfo, error) {
	return r.Resolver.Boiler.GetInfo(ctx)
}

// Sensor is the resolver for the sensor field.
func (r *queryResolver) Sensor(ctx context.Context, name string, position string) (*model.Measure, error) {
	from := time.Now().Add(-10 * time.Minute)
	to := time.Now()
	result, err := r.Resolver.Sensors[name+":"+position].Get(ctx, from, to)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[len(result)-1], nil
}

// SensorRange is the resolver for the sensorRange field.
func (r *queryResolver) SensorRange(ctx context.Context, name string, position string, from *time.Time, to *time.Time) ([]*model.Measure, error) {
	defaultFrom := time.Now().Add(-24 * time.Hour)
	defaultTo := time.Now()
	if from == nil {
		from = &defaultFrom
	}
	if to == nil {
		to = &defaultTo
	}
	return r.Resolver.Sensors[name+":"+position].Get(ctx, *from, *to)
}

// SwitchHistory is the resolver for the switchHistory field.
func (r *queryResolver) SwitchHistory(ctx context.Context, from *time.Time, to *time.Time) ([]*model.SwitchSample, error) {
	defaultFrom := time.Now().Add(-24 * time.Hour)
	defaultTo := time.Now()
	if from == nil {
		from = &defaultFrom
	}
	if to == nil {
		to = &defaultTo
	}
	return r.Resolver.Boiler.GetSwitchHistory(ctx, *from, *to)
}

// OverheatingProtectionHistory is the resolver for the overheatingProtectionHistory field.
func (r *queryResolver) OverheatingProtectionHistory(ctx context.Context, from *time.Time, to *time.Time) ([]*model.OverheatingProtectionSample, error) {
	defaultFrom := time.Now().Add(-24 * time.Hour)
	defaultTo := time.Now()
	if from == nil {
		from = &defaultFrom
	}
	if to == nil {
		to = &defaultTo
	}
	return r.Resolver.Boiler.GetOverheatingProtectionHistory(ctx, *from, *to)
}

// Boiler is the resolver for the boiler field.
func (r *subscriptionResolver) Boiler(ctx context.Context) (<-chan *model.BoilerInfo, error) {
	return r.Resolver.Boiler.Listen(ctx)
}

// Sensor is the resolver for the sensor field.
func (r *subscriptionResolver) Sensor(ctx context.Context, name string, position string) (<-chan *model.Measure, error) {
	return r.Resolver.Sensors[name+":"+position].Listen(ctx)
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
