package telegram

import(
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/4m4rOk/Mr-Proper/configuration"
)

var Bot, _ = tgbotapi.NewBotAPI(configuration.Config.Telegram.Token)
var UpdateConfig tgbotapi.UpdateConfig

func init() {
	Bot.Debug = configuration.Config.Telegram.Debug
	
	UpdateConfig = tgbotapi.NewUpdate(0)
	UpdateConfig.Timeout = 60
}