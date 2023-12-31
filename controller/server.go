//go:generate go run ./gqlgen.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"stupid-caldaia/controller/graph"
	"stupid-caldaia/controller/store"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {
	ctx := context.Background()
	config, err := store.LoadConfig()
	if err != nil {
		panic(err)
	}

	client, sensors, boiler := config.CreateObjects(context.Background())
	go store.TemperatureChangeController(ctx, boiler, sensors["temperatura:centrale"])
	go store.RuleEnforceController(ctx, boiler, sensors["temperatura:centrale"])
	go store.RuleTimingController(ctx, boiler)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Client: client, Sensors: sensors, Boiler: boiler}}))
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
	log.Panic(http.ListenAndServe(":"+port, nil))
}
