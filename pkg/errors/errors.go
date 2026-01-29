package errors

import "errors"

var (
	ErrNilDependency  = errors.New("nil dependency: ")
	ErrEntityNotFound = errors.New("entitie not found")
	ErrOpenState 	  = errors.New("circuit breaker is open")
)