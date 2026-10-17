package http

import (
	"google.golang.org/grpc/status"
)

func statusFromErr(err error) *status.Status {
	if st, ok := status.FromError(err); ok {
		return st
	}
	return status.New(13, err.Error()) // codes.Internal
}
