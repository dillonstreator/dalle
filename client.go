package dalle

import (
	"context"
	"io"
)

type Client interface {
	Generate(ctx context.Context, prompt string) (*Task, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
	Download(ctx context.Context, generationID string) (io.ReadCloser, error)
}
