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
	PermissionDenied ErrorCode = "permission_denied"
	Unimplemented    ErrorCode = "unimplemented"
	Internal         ErrorCode = "internal"
	RateLimited      ErrorCode = "rate_limited"
	Unavailable      ErrorCode = "unavailable"
	MissingArgument ErrorCode = "missing_argument"

	BadRoute  ErrorCode = "bad_route"
	Malformed ErrorCode = "malformed"
)

func newError(code ErrorCode, message string, statusCode int) GonzoError {
	return gerr{code: code, msg: message, statusCode: statusCode}
}

func InvalidArgumentError(msg string) GonzoError {
	return newError(InvalidArgument, msg, http.StatusBadRequest)
}

func MissingArgumentError(msg string) GonzoError {
	return newError(MissingArgument, msg, http.StatusBadRequest)
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

func PermissionDeniedError(msg string) GonzoError {
	return newError(PermissionDenied, msg, http.StatusForbidden)
}

func UnimplementedError(msg string) GonzoError {
	return newError(Unimplemented, msg, http.StatusNotImplemented)
}

func RateLimitedError(msg string) GonzoError {
	return newError(RateLimited, msg, http.StatusTooManyRequests)
}

func UnavailableError(msg string) GonzoError {
	return newError(Unavailable, msg, http.StatusServiceUnavailable)
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
