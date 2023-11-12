package bot

import (
	"fmt"
	"log"
	"strings"
	"suomi_news_channel/newsLog"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
)

var bot *tgbotapi.BotAPI

func Init(botToken string) {
	var err error
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is initialized")
}

func AskForApproval(adminChannelId int64, item *newsLog.EnhancedFeedItem) (*tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(adminChannelId, fmt.Sprintf("Постим?\n\n%s\n%s", item.Item.Title, item.Item.Link))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", fmt.Sprintf("yes:%s", item.ID)),
			tgbotapi.NewInlineKeyboardButtonData("Нет", fmt.Sprintf("no:%s", item.ID)),
		),
	)

	message, err := bot.Send(msg)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

func SubscribeForUpdates(wg *sync.WaitGroup, channelId int64, adminChannelId int64) {
	defer wg.Done()
	log.Println("[SubscribeForUpdates] START")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			handleBotUpdate(update, channelId, adminChannelId)
		}
	}
	log.Println("[SubscribeForUpdates] END")
}

// Handle updates coming from the admin channel (approve/reject posts, etc)
func handleBotUpdate(update tgbotapi.Update, channelId int64, adminChannelId int64) {
	if update.CallbackQuery != nil {
		isApproved := strings.Split(update.CallbackQuery.Data, ":")[0] == "yes"
		postId := strings.Split(update.CallbackQuery.Data, ":")[1]
		feedItem, err := newsLog.GetPostByID(postId)
		if err != nil {
			log.Println("Could not get post info from Redis")
			// TODO: Send error & resubmission request message to admin
		}
		if isApproved {
			PostPieceOfNews(channelId, &feedItem.Item)
			// TODO: actually verify that it was published instead of assuming it
			feedItem.Status = "published"
		} else {
			feedItem.Status = "rejected"
		}

		newsLog.SavePostToRedis(feedItem)
		NotifyAdminAboutPosting(update.CallbackQuery.Message.Chat.ID, isApproved)
		DeleteQuestionMessage(adminChannelId, feedItem.ApproveMessage)
	}
}

func PostPieceOfNews(channelId int64, item *gofeed.Item) {
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

	log.Println("Posted a piece of news")
}

func DeleteQuestionMessage(adminChannelId int64, message tgbotapi.Message) {
	deleteMsg := tgbotapi.NewDeleteMessage(adminChannelId, message.MessageID)

	_, err := bot.Send(deleteMsg)
	if err != nil {
		log.Println(err)
	}

	log.Println("Removing old approval question")
}

func NotifyAdminAboutPosting(channelId int64, approved bool) {
	replyText := "Ок, больше не буду спрашивать про эту новость"
	if approved {
		replyText = "Запостили, считаем лайки"
	}

	reply := tgbotapi.NewMessage(channelId, replyText)

	_, err := bot.Send(reply)
	if err != nil {
		log.Println(err)
	}

	log.Println("Admin was informed about an action")
}
