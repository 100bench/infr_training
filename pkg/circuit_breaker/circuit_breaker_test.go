package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_Concurrency(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()

			err := cb.Call(ctx, func() error {
				if id%2 == 0 {
					return errors.New("test error")
				}
				return nil
			})

			_ = err

			_ = cb.State()
		}(i)
	}

	wg.Wait()

	t.Log("Final state:", cb.State())
}