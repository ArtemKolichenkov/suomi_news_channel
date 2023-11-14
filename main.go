package main

import (
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	bot "suomi_news_channel/bot"
	newsLog "suomi_news_channel/newsLog"
	utils "suomi_news_channel/utils"
)

var scheduleIntervalMinutes = 10

func main() {
	var wg sync.WaitGroup
	initLog()
	botToken, channelId, adminChannelId, redisUrl, redisUsername, redisPassword := utils.CheckEnv()
	newsLog.InitRedisClient(redisUrl, redisUsername, redisPassword)
	bot.Init(botToken)

	// Run goroutine where bot listens for updates
	wg.Add(1)
	go bot.SubscribeForUpdates(&wg, channelId, adminChannelId)

	// Run goroutine that checks RSS feed every N minutes
	go proposeNewsEveryNSeconds(60*scheduleIntervalMinutes, adminChannelId)

	wg.Wait()
	bot.SendMessageToAdminChat(adminChannelId, "FATAL ERROR: Looks like bot stopped working, check logs.")
	log.Println("->> main END <<-")
}

func initLog() {
	log.SetOutput(os.Stdout) // todo move log into wrapper and log both in file and cli
}

func proposeNewsEveryNSeconds(n int, adminChannelId int64) {
	log.Println("[proposeNewsEveryNSeconds] First RSS fetch")
	proposeNews(adminChannelId) // Run once on start
	for range time.Tick(time.Second * time.Duration(n)) {
		dt := time.Now().Add(time.Minute * time.Duration(scheduleIntervalMinutes))
		log.Println("[proposeNewsEveryNSeconds] Scheduled RSS fetch started, next should be around", dt.Format("2006-01-02 15:04:05"))
		_, err := proposeNews(adminChannelId) // After N seconds run again, and so on forever
		if err != nil {
			log.Println("[proposeNewsEveryNSeconds] Error while fetching RSS feed:", err)
		}
		log.Println("[proposeNewsEveryNSeconds] Scheduled RSS fetch finished")
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
	// Sort the feed just in case RSS returns unsorted data (Newest first)
	sort.Slice(feed.Items, func(i, j int) bool {
		return feed.Items[i].PublishedParsed.After(*feed.Items[j].PublishedParsed)
	})
	log.Printf("[proposeNews] Fetched RSS feed with %d items", feed.Len())
	maxNewsCount := 2
	newsCount := 0
	for _, item := range feed.Items {
		enhancedItem := newsLog.EnhancedFeedItem{
			Item:   *item,
			Status: "suggested",
			ID:     utils.GetYleArticleId(item),
		}
		if newsLog.IsPublishedOrSuggested(&enhancedItem) {
			log.Printf("[proposeNews] Skipping Published/Suggested news item %s", enhancedItem.ID)
			continue
		}
		approvalMessage, approveErr := bot.AskForApproval(adminChannelId, &enhancedItem)
		if approveErr != nil {
			enhancedItem.Status = "suggestion_error"
			log.Printf("[proposeNews] Error while sending news item for approval %s", enhancedItem.ID)
		} else {
			log.Printf("[proposeNews] Successfully requested approval for news item %s", enhancedItem.ID)
			enhancedItem.ApproveMessage = *approvalMessage
		}
		err := newsLog.SavePostToRedis(&enhancedItem)
		if err != nil {
			log.Println("[proposeNews] Error while saving post to Redis, postId=", enhancedItem.ID, err)
		}
		log.Printf("[proposeNews] Successfully saved news item %s to Redis", enhancedItem.ID)

		newsCount++
		if newsCount >= maxNewsCount {
			log.Println("[proposeNews] Processed maxNewsCount of items, breaking")
			break
		}
	}
	log.Println("[proposeNews] Done")
	return newsCount, nil
}

func getFeed() (*gofeed.Feed, error) {
	fp := gofeed.NewParser()

	yliFeedURL := "https://feeds.yle.fi/uutiset/v1/recent.rss?publisherIds=YLE_NOVOSTI"

	feed, err := fp.ParseURL(yliFeedURL)
	return feed, err
}
