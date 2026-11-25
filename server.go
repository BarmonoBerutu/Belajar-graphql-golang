package main

import (
	"context"
	"log"
	"simple-userfollow-graphql/config"
	"simple-userfollow-graphql/directives"
	"simple-userfollow-graphql/graph"
	"simple-userfollow-graphql/middleware"
	"simple-userfollow-graphql/tools"

	"net/http"
	"os"
	"runtime/debug"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	config.ConnectDB()
	db := config.GetDB()
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	router := chi.NewRouter()
	router.Use(
		middleware.AuthMiddleware,
	)

	c := graph.Config{Resolvers: &graph.Resolver{}}
	c.Directives.IsLogin = directives.IsLogin

	srv := handler.New(graph.NewExecutableSchema(c))
	srv.SetRecoverFunc(func(ctx context.Context, err any) (userMessage error) {
		log.Printf("panic error: %+v\n", err)
		debug.PrintStack()

		return &gqlerror.Error{
			Message: "Internal server error",
			Extensions: map[string]any{
				"code": tools.GqlInternalError,
			},
		}
	})

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
