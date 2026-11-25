package directives

import (
	"context"
	"simple-userfollow-graphql/middleware"
	"simple-userfollow-graphql/tools"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func IsLogin(ctx context.Context, obj any, next graphql.Resolver) (res any, err error) {
	if middleware.AuthContext(ctx) == nil {
		return nil, &gqlerror.Error{
			Message: "Unauthenticated",
			Extensions: map[string]any{
				"code": tools.GqlUnauthenticated,
			},
		}
	}

	return next(ctx)
}
