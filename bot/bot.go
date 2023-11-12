package bot

import (
	"fmt"
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mmcdole/gofeed"
)

var logger *log.Logger
var bot *tgbotapi.BotAPI

func Init(botToken string){
    var err error
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is initialized")
}

func AskForApproval(adminChannelId int64, item *gofeed.Item) tgbotapi.Message {
    msg := tgbotapi.NewMessage(adminChannelId, fmt.Sprintf("Постим?\n\n%s\n%s", item.Title, item.Link))
    msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("Да", "yes"),
            tgbotapi.NewInlineKeyboardButtonData("Нет", "no"),
        ),
    )

    message, err := bot.Send(msg)
    if err != nil {
    fmt.Println(err)
        log.Fatal(err)
    }

	log.Println("Bot asked for approval")

    return message;
}

func GetUpdatesOnApprovals() []tgbotapi.Update {
    var result []tgbotapi.Update;

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates, _ := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.CallbackQuery != nil {
            result = append(result, update);

	        log.Println("Got updates on approval")

            return result
        }
    }

    return result;
}

func PostPieceOfNews(channelId int64, item *gofeed.Item){
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

func DeleteQuestionMessage(adminChannelId int64, message tgbotapi.Message){
    deleteMsg := tgbotapi.NewDeleteMessage(adminChannelId, message.MessageID);

    _, err := bot.Send(deleteMsg)
    if err != nil {
        log.Println(err)
    }

	log.Println("Removing old approval question")
}

func NotifyAdminAboutPosting(channelId int64, approved bool){
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

	log.Println("Admin was informed about an action")
}
