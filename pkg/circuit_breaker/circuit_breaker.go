package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed: return "CLOSED"
	case StateOpen: return "OPEN"
	case StateHalfOpen: return "HALF_OPEN"
	default: return "UNKNOWN"
	}
}

type CircuitBreaker struct {
	state         State
	failureCount  int
	maxFailures   int
	maxHalfOpenSuccesses int
	resetTimeout  time.Duration
	halfOpenTimer *time.Timer
	halfOpenSuccesses int
	mtx           sync.RWMutex
	onStateChange func(string)
}

func NewCircuitBreaker(maxFailures, maxHalfOpenSuccesses, halfOpenSuccesses int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		maxHalfOpenSuccesses: maxHalfOpenSuccesses,
		halfOpenSuccesses: halfOpenSuccesses,
	}
}

func (cb *CircuitBreaker) Call(ctx context.Context, operation func() error) error{
	cb.mtx.RLock()
	defer cb.mtx.RUnlock()
	switch cb.state {
    case StateOpen:
        return errors.New("circuit breaker is open")
	case StateHalfOpen:
		err := operation()
		if err != nil {
			cb.recordFailure()
			return err
		}
		cb.halfOpenSuccesses++
		if cb.halfOpenSuccesses >= cb.maxHalfOpenSuccesses {
			cb.reset()
		}
		return nil
    case StateClosed:
        err := operation()
        if err != nil {
            cb.recordFailure()
            return err
        }
        cb.reset()
        return nil
    }
    return nil

}

// recordFailure обновляет счетчик неудач и переключает состояние при необходимости
func (cb *CircuitBreaker) recordFailure() {
	cb.mtx.RLock()
	defer cb.mtx.RUnlock()
    cb.failureCount++
    if cb.failureCount >= cb.maxFailures {
        cb.transitionTo(StateOpen)
        cb.halfOpenTimer = time.AfterFunc(cb.resetTimeout, func() {
            cb.transitionTo(StateHalfOpen)
        })
    }
}

// reset сбрасывает счетчик неудач и переключает состояние на закрытое
func (cb *CircuitBreaker) reset() {
	cb.mtx.RLock()
	defer cb.mtx.RUnlock()
    if cb.halfOpenTimer != nil {
		cb.halfOpenTimer.Stop()
		cb.halfOpenTimer = nil
	}
	cb.failureCount = 0
	cb.transitionTo(StateClosed)
    
}

// transitionTo переключает состояние Circuit Breaker
func (cb *CircuitBreaker) transitionTo(state State) {
	cb.mtx.RLock()
	defer cb.mtx.RUnlock()
    cb.state = state
    if cb.onStateChange != nil {
        cb.onStateChange(state.String())
    }
}

func (cb *CircuitBreaker) State() string {
	cb.mtx.RLock()
	defer cb.mtx.RUnlock()
	return cb.state.String()
}