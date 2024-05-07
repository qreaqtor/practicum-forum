package redis

import (
	"encoding/json"
	"errors"
	"forum/internal/models"
	"time"

	"github.com/go-redis/redis"
)

var (
	tokenLifespan = 168 // in hours

	errNoKey = errors.New("key does not exists")
)

type redisStorage struct {
	client *redis.Client
}

func NewRedisStorage(client *redis.Client) *redisStorage {
	return &redisStorage{
		client: client,
	}
}

func (rs *redisStorage) Set(token string, author *models.Author) error {
	mkey := "token:" + token
	authorSerrialized, err := json.Marshal(author)
	if err != nil {
		return err
	}
	err = rs.client.Set(mkey, authorSerrialized, time.Hour*time.Duration(tokenLifespan)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rs *redisStorage) Get(token string) (*models.Author, error) {
	mkey := "token:" + token
	result := rs.client.Get(mkey)
	if result.Err() == redis.Nil {
		return nil, errNoKey
	}
	data, err := result.Bytes()
	if err != nil {
		return nil, err
	}
	author := &models.Author{}
	err = json.Unmarshal(data, author)
	if err != nil {
		return nil, err
	}
	return author, nil
}
