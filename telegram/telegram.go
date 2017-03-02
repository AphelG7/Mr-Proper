package telegram

import (
	"github.com/4m4rOk/Mr-Proper/configuration"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var Bot, _ = tgbotapi.NewBotAPI(configuration.Config.Telegram.Token)
var UpdateConfig tgbotapi.UpdateConfig

func init() {
	Bot.Debug = configuration.Config.Telegram.Debug

	UpdateConfig = tgbotapi.NewUpdate(0)
	UpdateConfig.Timeout = configuration.Config.Telegram.Timeout
}
