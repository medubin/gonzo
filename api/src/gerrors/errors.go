package gerrors

import (
	"net/http"
)

type ErrorCode string

const (
	InvalidArgument ErrorCode = "invalid_argument"
	NotFound        ErrorCode = "not_found"
	AlreadyExists   ErrorCode = "already_exists"
	Unauthenticated ErrorCode = "unauthenticated"
	Unimplemented   ErrorCode = "unimplemented"
	Internal        ErrorCode = "internal"

	BadRoute  ErrorCode = "bad_route"
	Malformed ErrorCode = "malformed"
)

func newError(code ErrorCode, message string, statusCode int) GonzoError {
	return gerr{code: code, msg: message, statusCode: statusCode}
}

func InvalidArgumentError(msg string) GonzoError {
	return newError(InvalidArgument, msg, http.StatusBadRequest)
}

func NotFoundError(msg string) GonzoError {
	return newError(NotFound, msg, http.StatusNotFound)
}

func AlreadyExistsError(msg string) GonzoError {
	return newError(AlreadyExists, msg, http.StatusConflict)
}

func UnauthenticatedError(msg string) GonzoError {
	return newError(Unauthenticated, msg, http.StatusUnauthorized)
}

func UnimplementedError(msg string) GonzoError {
	return newError(Unimplemented, msg, http.StatusNotImplemented)
}

func InternalError(msg string) GonzoError {
	return newError(Internal, msg, http.StatusInternalServerError)
}

// Internal

func BadRouteError(msg string) GonzoError {
	return newError(BadRoute, msg, http.StatusNotFound)
}

func MalformedError(msg string) GonzoError {
	return newError(Malformed, msg, http.StatusBadRequest)
}
