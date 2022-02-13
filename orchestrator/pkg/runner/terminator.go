package runner

import "context"

type Terminator interface {
	Terminate(ctx context.Context, id int) error
}
