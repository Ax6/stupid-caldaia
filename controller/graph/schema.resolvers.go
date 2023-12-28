package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.41

import (
	"context"
	"fmt"
	"strconv"
	"stupid-caldaia/controller/graph/model"
)

// SetSwitch is the resolver for the setSwitch field.
func (r *mutationResolver) SetSwitch(ctx context.Context, state *model.State) (model.State, error) {
	return model.Caldaia.Set(*state)
}

// Switch is the resolver for the switch field.
func (r *queryResolver) Switch(ctx context.Context) (*model.Switch, error) {
	return &model.Caldaia, nil
}

// OnTemperatureChange is the resolver for the onTemperatureChange field.
func (r *subscriptionResolver) OnTemperatureChange(ctx context.Context, position string) (<-chan float64, error) {
	key := "temperatura:" + position
	sub := r.Client.Subscribe(ctx, key)

	// Channel to send temperature updates to the GraphQL subscription
	temperatureUpdates := make(chan float64)

	// Goroutine to handle incoming messages from the Redis channel
	go func() {
		defer sub.Close()

		// Receive messages from the Redis channel
		for msg := range sub.Channel() {
			// Parse the temperature value from the message
			temperature, err := strconv.ParseFloat(msg.Payload, 64)
			if err != nil {
				// Handle error (e.g., log it)
				fmt.Printf("Error parsing temperature: %v\n", err)
				continue
			}

			// Send the temperature to the GraphQL subscription channel
			temperatureUpdates <- temperature
		}
		close(temperatureUpdates) // Close the channel when done
	}()

	return temperatureUpdates, nil
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