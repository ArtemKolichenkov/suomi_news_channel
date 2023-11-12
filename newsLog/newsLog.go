package newsLog

import (
	"context"
	"fmt"
	"time"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/mmcdole/gofeed"
)

var client *redis.Client
var ctx = context.Background()

// Initialize Redis client
func InitRedisClient() {
	client = redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Replace with your Redis server address
		Password: "",                // No password for local Redis
		DB:       0,                 // Default DB
	})
}

// IfPostWasPosted checks if a post was posted using Redis
func IfPostWasPosted(item *gofeed.Item) bool {
	key := fmt.Sprintf("post:%s", item.GUID) // Assuming GUID is unique for each post

	// Check if the key exists in Redis
	val, err := client.Exists(ctx, key).Result()
	if err != nil {
		log.Println("Error checking Redis:", err)
		return false
	}


	if (val == 1){
	    log.Println("Post was posted")

	    return true;
	}

	return false
}

// RememberPostWasPosted marks a post as posted in Redis
func RememberPostWasPosted(item *gofeed.Item) {
	key := fmt.Sprintf("post:%s", item.GUID) // Assuming GUID is unique for each post

	// Set the key in Redis with a TTL of 7 days (adjust as needed)
	err := client.SetEX(ctx, key, "1", 7*24*time.Hour).Err()
	if err != nil {
		log.Println("Error setting Redis key:", err)
	}

    log.Println("Remembered that post was posted")
}
