package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
    "github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

var adminChannelId int64
var logger *log.Logger
var bot *tgbotapi.BotAPI

func main() {
    initLog();

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Get the Telegram Bot Token from the environment variable
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	// Your Channel ID (replace with your channel ID)
	var channelId int64
	channelId, _ = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)

	adminChannelId, _ = strconv.ParseInt(os.Getenv("ADMIN_CHANNEL_ID"), 10, 64)

	// Initialize Telegram bot
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	feed := getFeed();

	// Check each feed item
	for _, item := range feed.Items {
		approvalMessage := askForApproval(item);

		// Wait for the user's reply
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates, _ := bot.GetUpdatesChan(u)

		for update := range updates {
			if update.CallbackQuery != nil {
				// User replied with "Yes" or "No"
				approved := false
				if update.CallbackQuery.Data == "yes" {
					postPieceOfNews(channelId, item);

                    approved = true;
				}

                notifyAdminAboutPosting(update.CallbackQuery.Message.Chat.ID, approved);
                deleteQuestionMessage(adminChannelId, approvalMessage);

				// Stop waiting for updates
				break
			}
		}

		break
	}

	// Handle potential errors
	if err != nil {
		log.Fatal(err)
	}
}

func postPieceOfNews(channelId int64, item *gofeed.Item){
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
}

func deleteQuestionMessage(adminChannelId int64, message tgbotapi.Message){
    deleteMsg := tgbotapi.NewDeleteMessage(adminChannelId, message.MessageID);

    _, err := bot.Send(deleteMsg)
    if err != nil {
        log.Println(err)
    }
}

func notifyAdminAboutPosting(channelId int64, approved bool){
    replyText := "Ок, больше не буду спрашивать про эту новость";
    if (approved){
        replyText = "Запостили, считаем лайки"
    }

    // Send a reply to the user
    reply := tgbotapi.NewMessage(channelId, replyText)

    _, err := bot.Send(reply)
    if err != nil {
        log.Println(err)
    }
}

func initLog(){
    logfile, err := os.Create("app.log")

    if err != nil {
        log.Fatal(err)
    }

    defer logfile.Close()
    log.SetOutput(logfile)
}

func askForApproval(item *gofeed.Item) tgbotapi.Message {
    msg := tgbotapi.NewMessage(adminChannelId, fmt.Sprintf("Постим?\n\n%s\n%s", item.Title, item.Link))
    msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("Да", "yes"),
            tgbotapi.NewInlineKeyboardButtonData("Нет", "no"),
        ),
    )

    message, err := bot.Send(msg)
    if err != nil {
        log.Fatal(err)
    }

    return message;
}

func getFeed() *gofeed.Feed {
	fp := gofeed.NewParser()

	yliFeedURL := "https://feeds.yle.fi/uutiset/v1/recent.rss?publisherIds=YLE_NOVOSTI"

	feed, err := fp.ParseURL(yliFeedURL)
	if err != nil {
		log.Fatal(err)
	}

	return feed;
}
