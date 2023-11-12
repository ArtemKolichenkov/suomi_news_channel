package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	botToken, channelId, adminChannelId, redisUrl, httpHost, httpPort := utils.CheckEnv()
	newsLog.InitRedisClient(redisUrl)
	bot.Init(botToken)

	wg.Add(1)
	go dummyHttpServer(httpHost, httpPort)

	// Run goroutine where bot listens for updates
	wg.Add(1)
	go bot.SubscribeForUpdates(&wg, channelId, adminChannelId)

	// Run goroutine that checks RSS feed every N minutes
	go proposeNewsEveryNSeconds(60*20, adminChannelId)

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
	maxNewsCount := 2
	newsCount := 0
	for _, item := range feed.Items {
		enhancedItem := newsLog.EnhancedFeedItem{
			Item:   *item,
			Status: "suggested",
			ID:     utils.GetYleArticleId(item),
		}
		if newsLog.IsPublishedOrSuggested(&enhancedItem) {
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

// This is needed purely so render.com marks deploy as successful, since it looks for open 80 port
// We can only use web server on render.com, because other services don't have $0 tier.
// At the same time we can use this server to get some info from the bot, e.g. Redis content
// TODO: obviously rework this for prod so it's not open for everyone (or come up with other solution without http server)
func dummyHttpServer(httpHost string, httpPort string) {
	http.HandleFunc("/redis", func(w http.ResponseWriter, r *http.Request) {
		posts, err := newsLog.GetAllPosts()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting posts from Redis: %s", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		jsonPosts, err := json.Marshal(posts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error marshalling posts: %s", err), http.StatusInternalServerError)
			return
		}
		w.Write(jsonPosts)
	})

	if err := http.ListenAndServe(httpHost+":"+httpPort, nil); err != nil {
		// TODO: probably shouldn't fatal here but gracefully restart server or send notification
		log.Fatal(err)
	}
}
