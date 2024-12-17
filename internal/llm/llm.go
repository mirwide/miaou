package llm

import (
	"context"
	"maps"

	"github.com/mirwide/miaou/internal/config"
	"github.com/ollama/ollama/api"
	ollama "github.com/ollama/ollama/api"
	"github.com/rs/zerolog/log"
)

type LLM struct {
	*ollama.Client
}

func NewLLM() (*LLM, error) {
	l, err := ollama.ClientFromEnvironment()
	return &LLM{l}, err
}

func (l *LLM) PullImages(ctx context.Context, models map[string]config.Model) error {
	for model := range maps.Values(models) {
		req := &api.PullRequest{
			Model: model.Name,
		}
		progressFunc := func(resp api.ProgressResponse) error {
			log.Debug().Str("name", model.Name).Str("status", resp.Status).Int64("total", resp.Total).Int64("completed", resp.Completed).Msg("pull progress")
			return nil
		}
		if err := l.Pull(ctx, req, progressFunc); err != nil {
			return err
		}
		log.Info().Str("name", model.Name).Msg("llm image pulled")
	}
	return nil
}
