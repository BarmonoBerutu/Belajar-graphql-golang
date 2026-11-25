package tools

import (
	"reflect"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	GqlUnauthenticated string = "UNAUTHENTICATED"
	GqlForbidden       string = "FORBIDDEN"
	GqlBadInput        string = "BAD_USER_INPUT"
	GqlNotFound        string = "NOT_FOUND"
	GqlConflict        string = "CONFLICT"
	GqlInternalError   string = "INTERNAL_SERVER_ERROR"
)

func NewGqlError(message string, code string) error {
	gqlErr := &gqlerror.Error{
		Message: message,
		Extensions: map[string]any{
			"code": code,
		},
	}
	return gqlErr
}

func ToGqlError(err error) error {
	switch reflect.TypeOf(err).String() {
	case "*gqlerror.Error":
		return err.(error)
	default:
		gqlErr := &gqlerror.Error{
			Message: err.Error(),
			Extensions: map[string]any{
				"code": GqlInternalError,
			},
		}
		return gqlErr
	}
}
