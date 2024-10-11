package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mirwide/tgbot/internal/config"
	ollama "github.com/ollama/ollama/api"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Storage struct {
	rdb *redis.Client
	cfg *config.Config
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})
	return &Storage{
		rdb: rdb,
		cfg: cfg,
	}, nil
}
func (s *Storage) GetMessages(chatID int64) []ollama.Message {

	ctx := context.Background()
	key := fmt.Sprintf("chat:%d:messages", chatID)

	textMessages, err := s.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		log.Error().Err(err).Msg("storage: problem get messages")
		return []ollama.Message{}
	}
	if len(textMessages) > s.cfg.Storege.History {
		s.rdb.LTrim(ctx, key, int64(len(textMessages)-s.cfg.Storege.History), -1)
	}
	var messages []ollama.Message
	for _, textMessage := range textMessages {
		var m ollama.Message
		if err := json.Unmarshal([]byte(textMessage), &m); err != nil {
			log.Error().Err(err).Msg("storage: problem unmarshal messages")
			continue
		}
		messages = append(messages, m)
	}
	return messages
}

func (s *Storage) SaveMessage(chatID int64, message ollama.Message) error {
	ctx := context.Background()
	key := fmt.Sprintf("chat:%d:messages", chatID)

	textMessage, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("storage: problem marshal")
		return err
	}
	if err := s.rdb.RPush(ctx, key, textMessage).Err(); err != nil {
		log.Error().Err(err).Msg("storage: problem save message")
		return err
	}

	if err := s.rdb.Expire(ctx, key, s.cfg.Storege.TTL).Err(); err != nil {
		log.Error().Err(err).Msg("storage: probles set expires")
		return err
	}
	return nil
}
