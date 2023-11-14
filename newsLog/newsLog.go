package newsLog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
)

var client *redis.Client
var ctx = context.Background()

// TODO: use "enums" for statuses instead of strings ("suggested", "published", "rejected", "suggestion_error")
type EnhancedFeedItem struct {
	Item           gofeed.Item
	ID             string
	Status         string
	ApproveMessage tgbotapi.Message
}

func InitRedisClient(redisUrl string, redisUsername string, redisPassword string) {
	client = redis.NewClient(&redis.Options{
		Addr:     redisUrl,      // Replace with your Redis server address
		Username: redisUsername, // "" for local Redis
		Password: redisPassword, // "" for local Redis
		DB:       0,             // Default DB
	})
}

func IsPublishedOrSuggested(item *EnhancedFeedItem) bool {
	item, err := GetPostByID(item.ID)
	if err != nil {
		// TODO: This actually might backfire, as this might be some fluke with Redis or unmarshling.
		// If we re-suggest based on this false, we might end up publishing same post twice
		return false
	}

	return item.Status == "published" || item.Status == "suggested"
}

func SavePostToRedis(item *EnhancedFeedItem) error {
	key := fmt.Sprintf("post:%s", item.ID)

	itemString, err := json.Marshal(item)
	if err != nil {
		return err
	}

	err = client.SetEX(ctx, key, itemString, 7*24*time.Hour).Err()
	return err
}

func GetPostByID(id string) (*EnhancedFeedItem, error) {
	key := fmt.Sprintf("post:%s", id)

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var item EnhancedFeedItem
	err = json.Unmarshal([]byte(val), &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}
