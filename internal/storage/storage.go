package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mirwide/miaou/internal/config"
	ollama "github.com/ollama/ollama/api"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	messageKey      = "messages:chat:%d"
	conversationKey = "conversation:chat:%d"
)

type Storage struct {
	rdb *redis.Client
	cfg *config.Config
}

type Conversation struct {
	Model string
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
		DB:   cfg.Storege.Db,
	})
	return &Storage{
		rdb: rdb,
		cfg: cfg,
	}, nil
}
func (s *Storage) GetMessages(chatID int64) []ollama.Message {

	ctx := context.Background()
	key := fmt.Sprintf(messageKey, chatID)

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

func (s *Storage) SaveMessage(chatID int64, message ollama.Message) {
	ctx := context.Background()
	key := fmt.Sprintf(messageKey, chatID)

	textMessage, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("storage: problem marshal")
	}
	if err := s.rdb.RPush(ctx, key, textMessage).Err(); err != nil {
		log.Error().Err(err).Msg("storage: problem save message")
	}

	if err := s.rdb.Expire(ctx, key, s.cfg.Storege.TTL).Err(); err != nil {
		log.Error().Err(err).Msg("storage: probles set expires")
	}
}

func (s *Storage) Clear(chatID int64) {
	ctx := context.Background()
	key := fmt.Sprintf(messageKey, chatID)
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		log.Error().Err(err).Msg("storage: problem clear messages")
	}
}

func (s *Storage) SaveConversation(chatID int64, conversation Conversation) error {
	ctx := context.Background()
	key := fmt.Sprintf(conversationKey, chatID)
	textConv, err := json.Marshal(conversation)
	if err != nil {
		log.Error().Err(err).Msg("storage: problem marshal")
		return err
	}
	if err := s.rdb.Set(ctx, key, textConv, s.cfg.Storege.TTL).Err(); err != nil {
		log.Error().Err(err).Msg("storage: problem save conversation")
		return err
	}
	return nil
}

func (s *Storage) GetConversation(chatID int64) (Conversation, error) {
	ctx := context.Background()
	key := fmt.Sprintf(conversationKey, chatID)
	conv, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return Conversation{}, err
	}
	if err != nil {
		log.Error().Err(err).Msg("storage: problem get conversation")
		return Conversation{}, err
	}
	var c Conversation
	if err := json.Unmarshal([]byte(conv), &c); err != nil {
		log.Error().Err(err).Msg("storage: problem unmarshal conversation")
		return Conversation{}, err
	}
	return c, nil
}
