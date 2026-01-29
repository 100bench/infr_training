package grpc_client

import (
	"context"
	pb "github.com/100bench/infr_training/pkg/proto/gen/notification/v1"
	"github.com/100bench/infr_training/todo_api/internal/usecases"
	cb "github.com/100bench/infr_training/pkg/circuit_breaker"
)

type NotifierWithCircuitBreaker struct{
	notifier usecases.Notifier
	cb *cb.CircuitBreaker
}

func NewNotifierWithCircuitBreaker(notifier usecases.Notifier, cb *cb.CircuitBreaker) *NotifierWithCircuitBreaker {
	return &NotifierWithCircuitBreaker{
		notifier: notifier,
		cb: cb,
	}
}

func (n *NotifierWithCircuitBreaker) Send(ctx context.Context, id, title string, notifyType pb.EventType) error{
	return n.cb.Call(ctx, func() error {
		return n.notifier.Send(ctx, id, title, notifyType)
	})
}