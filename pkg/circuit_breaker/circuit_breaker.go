package circuitbreaker

import (
	"context"
	"sync"
	"time"

	er "github.com/100bench/infr_training/pkg/errors"
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
	halfOpenAttempts    int
	maxHalfOpenAttempts    int
	resetTimeout  time.Duration
	halfOpenTimer *time.Timer
	halfOpenSuccesses int
	mtx           sync.Mutex
	onStateChange func(State)
}

func NewCircuitBreaker(maxFailures, maxHalfOpenSuccesses, maxHalfOpenAttempts int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		maxHalfOpenSuccesses: maxHalfOpenSuccesses,
		maxHalfOpenAttempts: maxHalfOpenAttempts,
	}
}

func (cb *CircuitBreaker) Call(ctx context.Context, operation func() error) error{
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cb.mtx.Lock()
	currentState := cb.state

	// 3. СПЕЦИАЛЬНАЯ ЛОГИКА ДЛЯ HALF_OPEN
	// Проблема: если 100 запросов приходят одновременно в HALF_OPEN,
	// мы не хотим их все пропускать к восстанавливающемуся сервису.
	// Решение: пропускаем только первые N (maxHalfOpenAttempts) запросов.
	if currentState == StateHalfOpen {
		// Проверяем: не превышен ли лимит попыток?
		if cb.halfOpenAttempts >= cb.maxHalfOpenAttempts {
			cb.mtx.Unlock()
			// Лимит превышен → отклоняем запрос сразу (Fail Fast)
			// Этот запрос НЕ доходит до сервиса!
			return er.ErrTooManyRequests
		}
		// Лимит НЕ превышен → увеличиваем счётчик попыток
		// Этот запрос разрешён к выполнению
		cb.halfOpenAttempts++
	}

	cb.mtx.Unlock()

	switch currentState {
	case StateOpen:
		return er.ErrOpenState

	case StateHalfOpen:
		err := operation()
		cb.mtx.Lock()
		if err != nil {
			cb.recordFailure()
			cb.mtx.Unlock()
			return err
		}

		cb.halfOpenSuccesses++
		if cb.halfOpenSuccesses >= cb.maxHalfOpenSuccesses {
			cb.reset()
		}

		cb.mtx.Unlock()
		return nil

	case StateClosed:
		err := operation()

		cb.mtx.Lock()
		if err != nil {
			cb.recordFailure()
		} else {
			cb.failureCount = 0
		}
		cb.mtx.Unlock()
		return err
	}

	return nil
}

// recordFailure обновляет счетчик неудач и переключает состояние при необходимости
func (cb *CircuitBreaker) recordFailure() {
    cb.failureCount++
    if cb.failureCount >= cb.maxFailures {
        cb.transitionTo(StateOpen)
		cb.halfOpenSuccesses = 0
        cb.halfOpenTimer = time.AfterFunc(cb.resetTimeout, func() {
			cb.mtx.Lock()
			defer cb.mtx.Unlock()
            cb.transitionTo(StateHalfOpen)
        })
    }
}

// reset сбрасывает счетчик неудач и переключает состояние на закрытое
func (cb *CircuitBreaker) reset() {
    if cb.halfOpenTimer != nil {
		cb.halfOpenTimer.Stop()
		cb.halfOpenTimer = nil
	}
	cb.failureCount = 0
	cb.halfOpenSuccesses = 0
	cb.transitionTo(StateClosed)
    
}

// transitionTo переключает состояние Circuit Breaker
// ⚠️ ВАЖНО: вызывающий должен держать lock!
func (cb *CircuitBreaker) transitionTo(state State) {
    cb.state = state

    // При переходе в HALF_OPEN сбрасываем счётчики
    // Это даёт возможность новым запросам попробовать (fresh start)
    if state == StateHalfOpen {
        cb.halfOpenAttempts = 0    // разрешаем новые попытки
        cb.halfOpenSuccesses = 0   // начинаем считать успехи заново
    }

    // Вызываем callback если есть (для логирования/метрик)
    if cb.onStateChange != nil {
        cb.onStateChange(state)
    }
}

func (cb *CircuitBreaker) State() State {
	cb.mtx.Lock()
	defer cb.mtx.Unlock()
	return cb.state
}
