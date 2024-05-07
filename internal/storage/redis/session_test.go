package redis

import (
	"forum/internal/models"
	"reflect"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestErrNoKey(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sessionRepo := NewRedisStorage(client)
	token := "token"

	_, err = sessionRepo.Get(token)
	if err != errNoKey {
		t.Errorf("want errNokey, but have %v", err)
	}
}

func TestSetAndGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sessionRepo := NewRedisStorage(client)

	token := "token"
	author := &models.Author{
		ID:       primitive.NewObjectID(),
		Username: "usertest",
	}

	err = sessionRepo.Set(token, author)
	if err != nil {
		t.Errorf("want error nil, but have %v", err)
	}

	authorStorage, err := sessionRepo.Get(token)
	if err != nil {
		t.Errorf("want error nil, but have %v", err)
	}
	if !reflect.DeepEqual(author, authorStorage) {
		t.Errorf("want %v, but have %v", author, authorStorage)
	}
}
