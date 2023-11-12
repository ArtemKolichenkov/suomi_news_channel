package main

import (
	"log"
	"os"

	"github.com/mmcdole/gofeed"

	bot "suomi_news_channel/bot"
	newsLog "suomi_news_channel/newsLog"
	"suomi_news_channel/utils"
)

func main() {
	initLog()

	log.Println("Starting script")

	botToken, channelId, adminChannelId, redisUrl := utils.CheckEnv()

	newsLog.InitRedisClient(redisUrl)

	bot.Init(botToken)

	feed := getFeed()

	maxNewsCount := 5
	newsCount := 0

	for _, item := range feed.Items {
		if newsLog.IfPostWasPosted(item) {
			continue
		}

		approvalMessage := bot.AskForApproval(adminChannelId, item)

		updates := bot.GetUpdatesOnApprovals()

		for _, update := range updates {
			approved := false

			if update.CallbackQuery.Data == "yes" {
				approved = true

				bot.PostPieceOfNews(channelId, item)
			}

			newsLog.RememberPostWasPosted(item)
			bot.NotifyAdminAboutPosting(update.CallbackQuery.Message.Chat.ID, approved)
			bot.DeleteQuestionMessage(adminChannelId, approvalMessage)

			// Stop waiting for updates
			break
		}

		newsCount++
		if newsCount >= maxNewsCount {
			break
		}
	}
}

func initLog() {
	log.SetOutput(os.Stdout) // todo move log into wrapper and log both in file and cli
}

func getFeed() *gofeed.Feed {
	fp := gofeed.NewParser()

	yliFeedURL := "https://feeds.yle.fi/uutiset/v1/recent.rss?publisherIds=YLE_NOVOSTI"

	feed, err := fp.ParseURL(yliFeedURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Imported feed")

	return feed
}
