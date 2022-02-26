package runner

import "context"

type LaunchInput struct {
	ID         uint64
	Owner      string
	Repository string
	Labels     []string
}

type Launcher interface {
	Launch(ctx context.Context, input *LaunchInput) error
}
