package public

import "context"

type ReadyChecker interface {
	Ready(ctx context.Context) error
}