package grpcerrors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsNotFound returns true if the specified error was created by NewNotFound.
// It supports wrapped errors and returns false when the error is nil.
func IsNotFound(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	if st.Code() != codes.NotFound {
		return false
	}
	return true
}

// IsNotFound returns true if the specified error was created by NewNotFound.
// It supports wrapped errors and returns false when the error is nil.
func InvalidArgument(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	if st.Code() != codes.InvalidArgument {
		return false
	}
	return true
}
