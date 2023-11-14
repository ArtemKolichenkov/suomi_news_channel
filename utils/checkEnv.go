package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Checks environmental variables
// Reports whether some of them are missing
func CheckEnv() (string, int64, int64, string, string, string) {
	appEnv := os.Getenv("APP_ENV")
	// In Railway env vars are present in runtime, not in .env file
	if appEnv != "PROD" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	channelId, _ := strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)
	if channelId == 0 {
		log.Fatal("CHANNEL_ID is not set")
	}

	adminChannelId, _ := strconv.ParseInt(os.Getenv("ADMIN_CHANNEL_ID"), 10, 64)
	if adminChannelId == 0 {
		log.Fatal("ADMIN_CHANNEL_ID is not set")
	}
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatal("REDIS_URL is not set")
	}
	var redisUsername string = ""
	var redisPassword string = ""
	if appEnv == "PROD" {
		// On localhost we don't need it
		redisUsername = os.Getenv("REDIS_USERNAME")
		if redisUsername == "" {
			log.Fatal("REDIS_USERNAME is not set")
		}
		redisPassword = os.Getenv("REDIS_PASSWORD")
		if redisPassword == "" {
			log.Fatal("REDIS_PASSWORD is not set")
		}
	}
	return botToken, channelId, adminChannelId, redisUrl, redisUsername, redisPassword
}
