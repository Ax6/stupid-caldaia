package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.41

import (
	"context"
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

// SetProgrammedInterval is the resolver for the setProgrammedInterval field.
func (r *mutationResolver) SetProgrammedInterval(ctx context.Context, id *string, start time.Time, duration time.Duration, targetTemp float64, repeatDays []model.DayOfWeek) (*model.ProgrammedInterval, error) {
	opt := &model.ProgrammedInterval{
		ID:         *id,
		Start:      start,
		Duration:   duration,
		TargetTemp: targetTemp,
		RepeatDays: repeatDays,
	}
	return r.Resolver.Boiler.SetProgrammedInterval(ctx, opt)
}

// DeleteProgrammedInterval is the resolver for the deleteProgrammedInterval field.
func (r *mutationResolver) DeleteProgrammedInterval(ctx context.Context, id string) (bool, error) {
	return r.Resolver.Boiler.DeleteProgrammedInterval(ctx, id)
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
