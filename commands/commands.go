package commands

import (
	"github.com/4m4rOk/Mr-Proper/telegram"
	"github.com/4m4rOk/Mr-Proper/functions"
	"github.com/4m4rOk/Mr-Proper/configuration"
	"github.com/4m4rOk/Mr-Proper/mongo"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"
	"html"
	"strconv"
)

func Id(message *tgbotapi.Message) {
	var response tgbotapi.MessageConfig

	if message.ReplyToMessage == nil {
		response = tgbotapi.NewMessage(message.Chat.ID, "<b>"+html.EscapeString(message.From.FirstName)+"'s ID:</b> "+strconv.Itoa(message.From.ID)+"\n<b>"+html.EscapeString(message.Chat.Title)+"'s ID:</b> "+strconv.FormatInt(message.Chat.ID, 10))
	} else {
		response = tgbotapi.NewMessage(message.Chat.ID, "<b>"+html.EscapeString(message.ReplyToMessage.From.FirstName)+"'s ID:</b> "+strconv.Itoa(message.ReplyToMessage.From.ID)+"\n<b>Message ID:</b> "+strconv.Itoa(message.ReplyToMessage.MessageID))
	}

	response.ReplyToMessageID = message.MessageID
	response.ParseMode = "HTML"
	telegram.Bot.Send(response)
}

func Link(message *tgbotapi.Message) {
	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	argument := message.CommandArguments()

	if argument == "" {
		var group mongo.Group
		collection.Find(bson.M{"_id": 0}).One(&group)

		var response tgbotapi.MessageConfig
		if group.Link != "" {
			response = tgbotapi.NewMessage(message.Chat.ID, html.EscapeString(message.From.FirstName)+", the link is:\n"+group.Link)
		} else {
			response = tgbotapi.NewMessage(message.Chat.ID, "I don't know the link. Sorry!")
		}
		response.ReplyToMessageID = message.MessageID
		telegram.Bot.Send(response)
	} else {
		user := functions.GetMember(message.Chat, message.From.ID)
		if user.IsAdministrator() || user.IsCreator() {
			functions.UpdateLink(message)
		}
	}
}

func Idle(message *tgbotapi.Message) {
	firstResponse := tgbotapi.NewMessage(message.Chat.ID, "Hang on... Let me see who has been naughty!")
	telegram.Bot.Send(firstResponse)

	argument := message.CommandArguments()

	var days float64
	days, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		days = 7.0
	}

	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var dusers []mongo.User
	collection.Find(bson.M{"Date": bson.M{"$lt": (message.Date - int(days*86400))}}).All(&dusers)

	var text string
	for _, duser := range dusers {
		member := functions.GetMember(message.Chat, duser.ID)

		if member.User.UserName != "" {
			text = text + html.EscapeString(member.User.FirstName) + " — @" + html.EscapeString(member.User.UserName) + "\n"
		} else {
			text = text + html.EscapeString(member.User.FirstName) + " — <i>" + strconv.Itoa(duser.ID) + "</i>\n"
		}
	}

	var secondResponse tgbotapi.MessageConfig
	if len(dusers) == 0 {
		secondResponse = tgbotapi.NewMessage(message.Chat.ID, "No one was inactive for more than "+strconv.FormatFloat(days, 'g', -1, 64)+" days. Happy?")
	} else if len(dusers) == 1 {
		secondResponse = tgbotapi.NewMessage(message.Chat.ID, "<b>This "+strconv.Itoa(len(dusers))+" user is inactive for at least "+strconv.FormatFloat(days, 'g', -1, 64)+" days:\n\n</b>"+text)
	} else {
		secondResponse = tgbotapi.NewMessage(message.Chat.ID, "<b>These "+strconv.Itoa(len(dusers))+" users are inactive for at least "+strconv.FormatFloat(days, 'g', -1, 64)+" days:</b>\n\n"+text)
	}

	secondResponse.ReplyToMessageID = message.MessageID
	secondResponse.ParseMode = "HTML"
	telegram.Bot.Send(secondResponse)
}

func Kick(message *tgbotapi.Message) {
	argument := message.CommandArguments()

	var response tgbotapi.MessageConfig

	var days float64
	days, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		days = 0
	}

	if days == 0 {
		response = tgbotapi.NewMessage(message.Chat.ID, message.From.FirstName+".... Nope. Nope. Nope. Nope.")
		response.ReplyToMessageID = message.MessageID
		telegram.Bot.Send(response)
		return
	}

	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var dusers []mongo.User
	collection.Find(bson.M{"Date": bson.M{"$lt": (message.Date - int(days*86400))}}).All(&dusers)

	for _, duser := range dusers {
		member := functions.GetMember(message.Chat, duser.ID)

		if member.IsCreator() {
			continue
		}

		var chatConfig = message.Chat.ChatConfig()
		var chatMemberConfig tgbotapi.ChatMemberConfig

		chatMemberConfig.ChatID = chatConfig.ChatID
		chatMemberConfig.SuperGroupUsername = chatConfig.SuperGroupUsername
		chatMemberConfig.UserID = duser.ID

		telegram.Bot.KickChatMember(chatMemberConfig)
		collection.Remove(bson.M{"_id": duser.ID})
		telegram.Bot.UnbanChatMember(chatMemberConfig)

	}

	if len(dusers) == 0 {
		response = tgbotapi.NewMessage(message.Chat.ID, "No one was inactive for more than "+strconv.FormatFloat(days, 'g', -1, 64)+" days. Happy?")
	} else {
		response = tgbotapi.NewMessage(message.Chat.ID, "Happy?")
	}

	response.ReplyToMessageID = message.MessageID
	telegram.Bot.Send(response)
}