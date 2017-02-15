package main

import (
	"github.com/4m4rOk/Mr-Proper/commands"
	"github.com/4m4rOk/Mr-Proper/functions"
	"github.com/4m4rOk/Mr-Proper/telegram"
	//"github.com/4m4rOk/Mr-Proper/configuration"
	"github.com/4m4rOk/Mr-Proper/mongo"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
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
			handlePrivate(update.Message)
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
		}
	}
}

func handleGroupMessage(message *tgbotapi.Message) {
	if message.NewChatMember != nil {
		if message.NewChatMember.ID == telegram.Bot.Self.ID {
			functions.NewGroup(message.Chat)
			return
		} else {
			functions.UpdateUser(message.NewChatMember, message.Chat, message.Date)
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

func handlePrivate(message *tgbotapi.Message) {
	response := tgbotapi.NewMessage(message.Chat.ID, "Hello dear creator. Make me admin in a supergroup of yours and I will show you how to rule properly.")
	telegram.Bot.Send(response)
}