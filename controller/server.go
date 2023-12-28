//go:generate go run ./gqlgen.go
package main

import (
	"log"
	"net/http"
	"os"

	"stupid-caldaia/controller/graph"
	"stupid-caldaia/controller/store"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/redis/go-redis/v9"
)

const defaultPort = "8080"

func main() {
	config, err := store.LoadConfig()
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(&config.Redis)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{ Client: client }}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
