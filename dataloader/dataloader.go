package dataloader

// Generate request-scoped loaders for user follow relations.

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"simple-userfollow-graphql/graph/model"
	"simple-userfollow-graphql/service"
	"time"
)

type loadersString string

const (
	loadersKey = loadersString("dataloaders")

	maxBatch = 100
	wait     = 1 * time.Millisecond
)

type Loaders struct {
	FollowersByUserID FollowersByUserIDLoader
	FollowingByUserID FollowingByUserIDLoader
}

func wrapFetchMany[K any, V any](fetch func(keys []K) ([][]*V, []error)) func(keys []K) ([][]*V, []error) {
	return func(keys []K) (results [][]*V, errs []error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("dataloader panic error: %+v\n", r)
				debug.PrintStack()

				errs = make([]error, len(keys))
				for i := range keys {
					errs[i] = fmt.Errorf("internal server error")
				}

				results = nil
			}
		}()

		return fetch(keys)
	}
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()

		loaders := &Loaders{
			FollowersByUserID: FollowersByUserIDLoader{
				maxBatch: maxBatch,
				wait:     wait,
				fetch: wrapFetchMany(func(ids []int) ([][]*model.User, []error) {
					items := make([][]*model.User, len(ids))

					usersByID, err := service.GetService().
						UserGetFollowersByIDs(requestCtx, ids)
					if err != nil {
						errs := make([]error, len(ids))
						for i := range errs {
							errs[i] = err
						}
						return nil, errs
					}

					for i, id := range ids {
						items[i] = usersByID[id]
						if items[i] == nil {
							items[i] = []*model.User{}
						}
					}

					return items, nil
				}),
			},

			FollowingByUserID: FollowingByUserIDLoader{
				maxBatch: maxBatch,
				wait:     wait,
				fetch: wrapFetchMany(func(ids []int) ([][]*model.User, []error) {
					items := make([][]*model.User, len(ids))

					usersByID, err := service.GetService().UserGetFollowingByIDs(requestCtx, ids)
					if err != nil {
						errs := make([]error, len(ids))
						for i := range errs {
							errs[i] = err
						}
						return nil, errs
					}

					for i, id := range ids {
						items[i] = usersByID[id]
						if items[i] == nil {
							items[i] = []*model.User{}
						}
					}

					return items, nil
				}),
			},
		}

		ctx := context.WithValue(requestCtx, loadersKey, loaders)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}
