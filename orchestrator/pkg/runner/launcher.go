package runner

import "context"

type LaunchInput struct {
	ID         int
	Owner      string
	Repository string
	Labels     []string
}

type Launcher interface {
	Launch(ctx context.Context, input *LaunchInput) error
}
