package newsLog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
)

var client *redis.Client
var ctx = context.Background()

type EnhancedFeedItem struct {
	Item           gofeed.Item
	ID             string
	Status         string
	ApproveMessage tgbotapi.Message
}

func InitRedisClient(redisUrl string) {
	client = redis.NewClient(&redis.Options{
		Addr:     redisUrl, // Replace with your Redis server address
		Password: "",       // No password for local Redis
		DB:       0,        // Default DB
	})
}

func IsPublished(item *EnhancedFeedItem) bool {
	item, err := GetPostByID(item.ID)
	if err != nil {
		// TODO: This actually might backfire. If we re-suggest based on this false, we might end up publishing same post twice
		log.Println("[GetPostByID] Error getting Redis key:", err)
		return false
	}

	return item.Status == "published"
}

func SavePostToRedis(item *EnhancedFeedItem) {
	key := fmt.Sprintf("post:%s", item.ID)

	itemString, err := json.Marshal(item)
	if err != nil {
		log.Println("Error marshalling item:", err)
	}

	err = client.SetEX(ctx, key, itemString, 7*24*time.Hour).Err()
	if err != nil {
		log.Println("Error setting Redis key:", err)
	}
}

func GetPostByID(id string) (*EnhancedFeedItem, error) {
	key := fmt.Sprintf("post:%s", id)

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Println("[GetPostByID] Error getting Redis key:", err)
		return nil, err
	}

	var item EnhancedFeedItem
	err = json.Unmarshal([]byte(val), &item)
	if err != nil {
		log.Println("[GetPostByID] Error unmarshalling item:", err)
		return nil, err
	}

	return &item, nil
}
