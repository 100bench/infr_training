package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	er "github.com/100bench/infr_training/pkg/errors"
)

// TestCircuitBreaker_Basic проверяет базовые переходы состояний
func TestCircuitBreaker_Basic(t *testing.T) {
	// Создаём CB: 3 ошибки → OPEN, в HALF_OPEN нужно 2 успеха, лимит попыток = 2
	// ⚠️ maxHalfOpenAttempts должен быть >= maxHalfOpenSuccesses!
	cb := NewCircuitBreaker(3, 2, 2, 100*time.Millisecond)

	ctx := context.Background()

	// 1. Начальное состояние: CLOSED
	if cb.State() != StateClosed {
		t.Errorf("Expected CLOSED, got %v", cb.State())
	}

	// 2. Делаем 3 ошибки подряд
	for i := 0; i < 3; i++ {
		err := cb.Call(ctx, func() error {
			return errors.New("test error")
		})
		if err == nil {
			t.Error("Expected error, got nil")
		}
	}

	// 3. Теперь должны быть в OPEN
	if cb.State() != StateOpen {
		t.Errorf("Expected OPEN after 3 failures, got %v", cb.State())
	}

	// 4. Попытка вызвать в OPEN → сразу ошибка
	err := cb.Call(ctx, func() error {
		t.Error("Operation should not be called in OPEN state!")
		return nil
	})
	if !errors.Is(err, er.ErrOpenState) {
		t.Errorf("Expected ErrOpenState, got %v", err)
	}

	// 5. Ждём timeout → переход в HALF_OPEN
	time.Sleep(150 * time.Millisecond)

	if cb.State() != StateHalfOpen {
		t.Errorf("Expected HALF_OPEN after timeout, got %v", cb.State())
	}

	// 6. В HALF_OPEN делаем 2 успешных запроса → CLOSED
	for i := 0; i < 2; i++ {
		err := cb.Call(ctx, func() error {
			return nil // успех
		})
		if err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}
	}

	// 7. Должны вернуться в CLOSED
	if cb.State() != StateClosed {
		t.Errorf("Expected CLOSED after 2 successes in HALF_OPEN, got %v", cb.State())
	}
}

// TestCircuitBreaker_HalfOpenLimit проверяет лимит запросов в HALF_OPEN
func TestCircuitBreaker_HalfOpenLimit(t *testing.T) {
	// maxHalfOpenAttempts = 1 → только 1 запрос разрешён в HALF_OPEN
	cb := NewCircuitBreaker(3, 2, 1, 50*time.Millisecond)
	ctx := context.Background()

	// Переводим в OPEN
	for i := 0; i < 3; i++ {
		cb.Call(ctx, func() error { return errors.New("error") })
	}

	// Ждём перехода в HALF_OPEN
	time.Sleep(100 * time.Millisecond)

	if cb.State() != StateHalfOpen {
		t.Fatalf("Expected HALF_OPEN, got %v", cb.State())
	}

	// Первый запрос → должен пройти
	err := cb.Call(ctx, func() error {
		return nil
	})
	if err != nil {
		t.Errorf("First request in HALF_OPEN should succeed, got error: %v", err)
	}

	// Второй запрос → должен быть отклонён (лимит = 1)
	err = cb.Call(ctx, func() error {
		t.Error("Second request should not execute operation!")
		return nil
	})
	if !errors.Is(err, er.ErrTooManyRequests) {
		t.Errorf("Expected ErrTooManyRequests, got %v", err)
	}
}

// TestCircuitBreaker_Concurrency проверяет thread-safety
func TestCircuitBreaker_Concurrency(t *testing.T) {
	cb := NewCircuitBreaker(5, 2, 3, 100*time.Millisecond)
	ctx := context.Background()

	var wg sync.WaitGroup

	// Запускаем 100 горутин одновременно
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Половина успешных, половина с ошибками
			err := cb.Call(ctx, func() error {
				time.Sleep(1 * time.Millisecond) // имитация работы
				if id%2 == 0 {
					return errors.New("test error")
				}
				return nil
			})

			_ = err

			// Читаем state одновременно из разных горутин
			_ = cb.State()
		}(i)
	}

	wg.Wait()

	t.Logf("Final state: %v", cb.State())
	// Тест просто проверяет что нет race conditions
	// Запусти: go test -race
}

// TestCircuitBreaker_ContextCancellation проверяет отмену контекста
func TestCircuitBreaker_ContextCancellation(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 1, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем сразу

	err := cb.Call(ctx, func() error {
		t.Error("Operation should not be called with cancelled context!")
		return nil
	})

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}
