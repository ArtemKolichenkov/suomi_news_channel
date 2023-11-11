package main

import (
	"fmt"
	"log"
	"os"
// 	"time"
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
    "github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Get the Telegram Bot Token from the environment variable
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	// Your Channel ID (replace with your channel ID)
	var channelId int64
	channelId, _ = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)

// 	var adminChannelId int64
// 	adminChannelId, _ = strconv.ParseInt(os.Getenv("ADMIN_CHANNEL_ID"), 10, 64)

	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	feed := getFeed();

	// Check each feed item
	for _, item := range feed.Items {
		// Send a message to the channel for approval
// 		msg := tgbotapi.NewMessage(adminChannelId, fmt.Sprintf("Do you want to post this news?\n\n%s\n%s", item.Title, item.Link))
// 		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
// 			tgbotapi.NewInlineKeyboardRow(
// 				tgbotapi.NewInlineKeyboardButtonData("Yes", "yes"),
// 				tgbotapi.NewInlineKeyboardButtonData("No", "no"),
// 			),
// 		)
//
// 		// Send the message
// 		message, err := bot.Send(msg)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
//
// 		// Wait for a reply for a limited time
// 		go func() {
// 			time.Sleep(60 * time.Second) // Adjust as needed
// 			// If no reply received, delete the message
// 			deleteMsg := tgbotapi.NewDeleteMessage(adminChannelId, message.MessageID)
// 			_, err := bot.Send(deleteMsg)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		}()
//
// 		// Wait for the user's reply
// 		u := tgbotapi.NewUpdate(0)
// 		u.Timeout = 60
// 		updates, err := bot.GetUpdatesChan(u)
//
// 		for update := range updates {
// 			if update.CallbackQuery != nil {
// 				// User replied with "Yes" or "No"
// 				replyText := ""
// 				if update.CallbackQuery.Data == "yes" {
					// Post the news to the channel
// 					replyText = fmt.Sprintf("Approved! Posting:\n\n%s\n%s", item.Title, item.Link)
					postMessage := tgbotapi.NewMessage(channelId, fmt.Sprintf("%s\n%s", item.Title, item.Link))
					postMessage.ParseMode = tgbotapi.ModeHTML
					postMessage.DisableWebPagePreview = false

					// Enable instant view feature
// 					postMessage.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
// 						tgbotapi.NewInlineKeyboardRow(
// 							tgbotapi.NewInlineKeyboardButtonSwitch("Read More", item.Link),
// 						),
// 					)

					_, err := bot.Send(postMessage)
					if err != nil {
						log.Println(err)
					}
// 				} else {
// 					replyText = "News not posted."
// 				}
//
// 				// Send a reply to the user
// 				reply := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, replyText)
// 				_, err := bot.Send(reply)
// 				if err != nil {
// 					log.Println(err)
// 				}

				// Stop waiting for updates
				break
// 			}
// 		}
	}

	// Handle potential errors
	if err != nil {
		log.Fatal(err)
	}
}

func getFeed() *gofeed.Feed {
	// Initialize RSS feed parser
	fp := gofeed.NewParser()

	// URL of the Yli.fi RSS feed
	yliFeedURL := "https://feeds.yle.fi/uutiset/v1/recent.rss?publisherIds=YLE_NOVOSTI"

	// Fetch the RSS feed
	feed, err := fp.ParseURL(yliFeedURL)
	if err != nil {
		log.Fatal(err)
	}

	return feed;
}
