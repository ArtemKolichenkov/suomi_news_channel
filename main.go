package main

import (
	"log"
	"os"
	"strconv"

    "github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"

	bot "suomi_news_channel/bot"
)

var adminChannelId int64

func main() {
    initLog();

    log.Println("Starting script");

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

	bot.Init(botToken);

	feed := getFeed();

	// Check each feed item
	for _, item := range feed.Items {
		approvalMessage := bot.AskForApproval(adminChannelId, item);

		updates := bot.GetUpdatesOnApprovals();

		for _, update := range updates {
            approved := false

            if update.CallbackQuery.Data == "yes" {
                approved = true;

                bot.PostPieceOfNews(channelId, item);
            }

            bot.NotifyAdminAboutPosting(update.CallbackQuery.Message.Chat.ID, approved);
            bot.DeleteQuestionMessage(adminChannelId, approvalMessage);

            // Stop waiting for updates
            break
		}

		break; // get only one message
	}

	// Handle potential errors
	if err != nil {
		log.Fatal(err)
	}
}

func initLog(){
    file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

    if err != nil {
        log.Fatal(err)
    }

    log.SetOutput(file);
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
