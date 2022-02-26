package runner

import "context"

type Terminator interface {
	Terminate(ctx context.Context, id uint64) error
}
