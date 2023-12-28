//go:generate go run ./gqlgen.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"stupid-caldaia/controller/graph"
	"stupid-caldaia/controller/graph/model"
	"stupid-caldaia/controller/store"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/redis/go-redis/v9"
)

const defaultPort = "8080"

func makeResolverDependencies(ctx context.Context, config store.Config) (*redis.Client, map[string]*model.Sensor) {
	client := redis.NewClient(&config.Redis)
	sensors := make(map[string]*model.Sensor)
	for _, sensorOptions := range config.Sensors {
		sensor, err := model.NewSensor(ctx, client, &sensorOptions)
		if err != nil {
			panic(err)
		}
		sensors[sensor.Id] = sensor
	}
	return client, sensors
}

func main() {
	config, err := store.LoadConfig()
	if err != nil {
		panic(err)
	}

	client, sensors := makeResolverDependencies(context.Background(), config)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Client: client, Sensors: sensors}}))
	srv.AddTransport(transport.SSE{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.Use(extension.Introspection{})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", c.Handler(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
