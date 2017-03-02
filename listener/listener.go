package main

import (
	"github.com/4m4rOk/Mr-Proper/commands"
	"github.com/4m4rOk/Mr-Proper/configuration"
	"github.com/4m4rOk/Mr-Proper/functions"
	"github.com/4m4rOk/Mr-Proper/mongo"
	"github.com/4m4rOk/Mr-Proper/telegram"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
)

func main() {
	defer mongo.Database.Close()

	updates, err := telegram.Bot.GetUpdatesChan(telegram.UpdateConfig)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Chat.IsPrivate() {
			if update.Message.IsCommand() {
				handlePrivateCommand(update.Message)
			} else {
				handlePrivateMessage(update.Message)
			}
		} else if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			if update.Message.IsCommand() {
				handleGroupCommand(update.Message)
			} else {
				handleGroupMessage(update.Message)
			}
		}
	}
}

func handleGroupCommand(message *tgbotapi.Message) {
	user := functions.GetMember(message.Chat, message.From.ID)

	if message.Command() == "link" {
		commands.Link(message)
	} else if message.Command() == "id" {
		commands.Id(message)
	}

	if user.IsAdministrator() || user.IsCreator() {
		if message.Command() == "idle" {
			commands.Idle(message)
		} else if message.Command() == "kick" {
			if message.Chat.IsSuperGroup() == false {
				response := tgbotapi.NewMessage(message.Chat.ID, "I can only kick in supergroups, sorry!")
				telegram.Bot.Send(response)
			} else {
				commands.Kick(message)
			}
		} else if message.Command() == "autoidle" {
			commands.AutoIdle(message)
		} else if message.Command() == "autokick" {
			if message.Chat.IsSuperGroup() == false {
				response := tgbotapi.NewMessage(message.Chat.ID, "I can only kick in supergroups, sorry!")
				telegram.Bot.Send(response)
			} else {
				commands.AutoKick(message)
			}
		}
	}
}

func handleGroupMessage(message *tgbotapi.Message) {
	if message.NewChatMember != nil {
		if message.NewChatMember.ID == telegram.Bot.Self.ID {
			functions.NewGroup(message.Chat)
			return
		} else {
			if message.NewChatMember.UserName != "" {
				username := strings.ToLower(message.NewChatMember.UserName)
				if strings.HasSuffix(username, "bot") == false {
					functions.UpdateUser(message.NewChatMember, message.Chat, message.Date)
				}
			}
			return
		}
	}
	if message.LeftChatMember != nil {
		if message.LeftChatMember.ID == telegram.Bot.Self.ID {
			functions.DeleteGroup(message.Chat)
			return
		} else {
			functions.DeleteUser(message.LeftChatMember, message.Chat)
			return
		}
	}
	functions.UpdateUser(message.From, message.Chat, message.Date)
}

func handlePrivateCommand(message *tgbotapi.Message) {
	for _, admin := range configuration.Config.Telegram.Admins {
		if admin == message.From.ID {
			if message.Command() == "groupslist" {
				commands.GroupsList(message)
			} else {
				handlePrivateMessage(message)
			}
		} else {
			handlePrivateMessage(message)
		}
	}
}

func handlePrivateMessage(message *tgbotapi.Message) {
	response := tgbotapi.NewMessage(message.Chat.ID, "Hello dear creator. Make me admin in a supergroup of yours and I will show you how to rule properly.")
	telegram.Bot.Send(response)
}
