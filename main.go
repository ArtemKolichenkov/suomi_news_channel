package main

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	bot "suomi_news_channel/bot"
	newsLog "suomi_news_channel/newsLog"
	utils "suomi_news_channel/utils"
)

func main() {
	var wg sync.WaitGroup
	initLog()
	botToken, channelId, adminChannelId, redisUrl := utils.CheckEnv()
	newsLog.InitRedisClient(redisUrl)
	bot.Init(botToken)

	// Run goroutine where bot listens for updates
	wg.Add(1)
	go bot.SubscribeForUpdates(&wg, channelId, adminChannelId)

	// Run goroutine that checks RSS feed every N minutes
	go proposeNewsEveryNSeconds(20, adminChannelId)

	wg.Wait()
	log.Println("->> main END <<-")
}

func initLog() {
	log.SetOutput(os.Stdout) // todo move log into wrapper and log both in file and cli
}

func proposeNewsEveryNSeconds(n int, adminChannelId int64) {
	proposeNews(adminChannelId) // Run once on start
	for range time.Tick(time.Second * time.Duration(n)) {
		proposeNews(adminChannelId) // After N seconds run again, and so on forever
	}
}

// Saves YLE posts into Redis with status "suggested" & pushes them to admin tg channel for approval
func proposeNews(adminChannelId int64) (int, error) {
	feed, err := getFeed()
	if err != nil {
		log.Println("[proposeNews] Error while fetching RSS feed:", err)
		// TODO: Send error notification to admin channel
		return 0, err
	}
	maxNewsCount := 1
	newsCount := 0
	for _, item := range feed.Items {
		enhancedItem := newsLog.EnhancedFeedItem{
			Item:   *item,
			Status: "suggested",
			ID:     utils.GetYleArticleId(item),
		}
		if newsLog.IsPublished(&enhancedItem) {
			continue
		}
		approvalMessage, approveErr := bot.AskForApproval(adminChannelId, &enhancedItem)
		if approveErr != nil {
			enhancedItem.Status = "suggestion_error"
		} else {
			enhancedItem.ApproveMessage = *approvalMessage
		}
		newsLog.SavePostToRedis(&enhancedItem)

		newsCount++
		if newsCount >= maxNewsCount {
			break
		}
	}
	return newsCount, nil
}

func getFeed() (*gofeed.Feed, error) {
	fp := gofeed.NewParser()

	yliFeedURL := "https://feeds.yle.fi/uutiset/v1/recent.rss?publisherIds=YLE_NOVOSTI"

	feed, err := fp.ParseURL(yliFeedURL)
	return feed, err
}
