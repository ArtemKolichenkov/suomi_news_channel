package main

import (
	"log"
	"os"
	"strconv"

    "github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"

	bot "suomi_news_channel/bot"
	newsLog "suomi_news_channel/newsLog"
)

var adminChannelId int64

func main() {
    initLog();

    log.Println("Starting script");

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	var channelId int64
	channelId, _ = strconv.ParseInt(os.Getenv("CHANNEL_ID"), 10, 64)

	adminChannelId, _ = strconv.ParseInt(os.Getenv("ADMIN_CHANNEL_ID"), 10, 64)

	newsLog.InitRedisClient();

	bot.Init(botToken);

	feed := getFeed();

	maxNewsCount := 5;
	newsCount := 0;

	for _, item := range feed.Items {
        if (newsLog.IfPostWasPosted(item)){
            continue;
        }

		approvalMessage := bot.AskForApproval(adminChannelId, item);

		updates := bot.GetUpdatesOnApprovals();

		for _, update := range updates {
            approved := false

            if update.CallbackQuery.Data == "yes" {
                approved = true;

                bot.PostPieceOfNews(channelId, item);
            }

            newsLog.RememberPostWasPosted(item)
            bot.NotifyAdminAboutPosting(update.CallbackQuery.Message.Chat.ID, approved);
            bot.DeleteQuestionMessage(adminChannelId, approvalMessage);

            // Stop waiting for updates
            break
		}

        newsCount++;
        if (newsCount >= maxNewsCount){
		    break;
		}
	}

	// Handle potential errors
	if err != nil {
		log.Fatal(err)
	}
}

func initLog(){
//     file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
//
//     if err != nil {
//         log.Fatal(err)
//     }
//    log.SetOutput(file);


    log.SetOutput(os.Stdout); // todo move log into wrapper and log both in file and cli
}

func getFeed() *gofeed.Feed {
	fp := gofeed.NewParser()

	yliFeedURL := "https://feeds.yle.fi/uutiset/v1/recent.rss?publisherIds=YLE_NOVOSTI"

	feed, err := fp.ParseURL(yliFeedURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Imported feed")

	return feed;
}
