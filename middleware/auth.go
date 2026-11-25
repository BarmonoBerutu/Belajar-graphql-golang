package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"simple-userfollow-graphql/tools"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var CtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type User struct {
	ID int `json:"id"`
}

func returnGraphQLError(w http.ResponseWriter, statusCode int, message string, code string) {
	resp := graphql.Response{
		Errors: gqlerror.List{
			&gqlerror.Error{
				Message: message,
				Extensions: map[string]any{
					"code": code,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&resp)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("Authorization")
		if authToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		authTokens := strings.Split(authToken, " ")
		if authTokens[0] != "Bearer" {
			returnGraphQLError(w, http.StatusUnauthorized, "Unauthenticated", tools.GqlUnauthenticated)
			return
		}

		jwtToken, err := tools.TokenValidate(authTokens[1])
		if err != nil {
			returnGraphQLError(w, http.StatusUnauthorized, "Unauthenticated", tools.GqlUnauthenticated)
			return
		}

		claims, ok := jwtToken.Claims.(*tools.JwtClaim)
		if !ok || !jwtToken.Valid {
			returnGraphQLError(w, http.StatusUnauthorized, "Unauthenticated", tools.GqlUnauthenticated)
			return
		}

		ctx := context.WithValue(r.Context(), CtxKey, &User{
			ID: claims.ID,
		})

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func AuthContext(ctx context.Context) *User {
	raw, _ := ctx.Value(CtxKey).(*User)
	return raw
}
