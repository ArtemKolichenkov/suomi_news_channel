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

// TODO: use "enums" for statuses instead of strings ("suggested", "published", "rejected", "suggestion_error")
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

func IsPublishedOrSuggested(item *EnhancedFeedItem) bool {
	item, err := GetPostByID(item.ID)
	if err != nil {
		// TODO: This actually might backfire. If we re-suggest based on this false, we might end up publishing same post twice
		log.Println("[GetPostByID] Error getting Redis key:", err)
		return false
	}

	return item.Status == "published" || item.Status == "suggested"
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
		log.Println("[GetPostByID] Error getting Redis key:", key, "error:", err)
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

// Note: only used for http API now
func GetAllPosts() (map[string]interface{}, error) {
	// TODO: Make it more type-safe, interface can screw things up down the line
	results := make(map[string]interface{})
	iter := client.Scan(ctx, 0, "post:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		value, err := client.Get(ctx, key).Result()
		if err != nil {
			results[key] = "error: not found in Redis"
		} else {
			var item EnhancedFeedItem
			err = json.Unmarshal([]byte(value), &item)
			if err != nil {
				results[key] = "error: failed unmarshalling"
			} else {
				results[key] = item
			}
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
