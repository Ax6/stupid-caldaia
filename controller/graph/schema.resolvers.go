package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.41

import (
	"context"
	"fmt"
	"stupid-caldaia/controller/graph/model"
	"time"
)

// UpdateBoiler is the resolver for the updateBoiler field.
func (r *mutationResolver) UpdateBoiler(ctx context.Context, config model.BoilerInput) (*model.BoilerInfo, error) {
	if config.State != nil {
		_, err := r.Boiler.Switch(ctx, *config.State)
		if err != nil {
			return nil, err
		}
	}

	if config.MinTemp != nil {
		_, err := r.Boiler.SetMinTemp(ctx, *config.MinTemp)
		if err != nil {
			return nil, err
		}
	}

	if config.MaxTemp != nil {
		_, err := r.Boiler.SetMaxTemp(ctx, *config.MaxTemp)
		if err != nil {
			return nil, err
		}
	}

	if config.TargetTemp != nil {
		_, err := r.Boiler.SetTargetTemp(ctx, *config.TargetTemp)
		if err != nil {
			return nil, err
		}
	}

	// if config.ProgrammedIntervals != nil {
	// 	_, err = boilerStore.SetProgrammedIntervals(ctx, config.ProgrammedIntervals)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return r.Resolver.Boiler.GetInfo(ctx)
}

// AddProgrammedInterval is the resolver for the addProgrammedInterval field.
func (r *mutationResolver) AddProgrammedInterval(ctx context.Context, interval model.ProgrammedIntervalInput) (*model.BoilerInfo, error) {
	panic(fmt.Errorf("not implemented: AddProgrammedInterval - addProgrammedInterval"))
}

// RemoveProgrammedInterval is the resolver for the removeProgrammedInterval field.
func (r *mutationResolver) RemoveProgrammedInterval(ctx context.Context, id string) (*model.BoilerInfo, error) {
	panic(fmt.Errorf("not implemented: RemoveProgrammedInterval - removeProgrammedInterval"))
}

// Boiler is the resolver for the boiler field.
func (r *queryResolver) Boiler(ctx context.Context) (*model.BoilerInfo, error) {
	return r.Resolver.Boiler.GetInfo(ctx)
}

// Sensor is the resolver for the sensor field.
func (r *queryResolver) Sensor(ctx context.Context, name string, position string) (*model.Measure, error) {
	result, err := r.Sensors[name+":"+position].Get(ctx, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

// SensorRange is the resolver for the sensorRange field.
func (r *queryResolver) SensorRange(ctx context.Context, name string, position string, from *time.Time, to *time.Time) ([]*model.Measure, error) {
	return r.Resolver.Sensors[name+":"+position].Get(ctx, from, to)
}

// Boiler is the resolver for the boiler field.
func (r *subscriptionResolver) Boiler(ctx context.Context) (<-chan *model.BoilerInfo, error) {
	return r.Resolver.Boiler.Listen(ctx)
}

// Sensor is the resolver for the sensor field.
func (r *subscriptionResolver) Sensor(ctx context.Context, name string, position string) (<-chan *model.Measure, error) {
	return r.Sensors[name+":"+position].Listen(ctx)
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
