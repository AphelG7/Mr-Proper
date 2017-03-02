package commands

import (
	"github.com/4m4rOk/Mr-Proper/configuration"
	"github.com/4m4rOk/Mr-Proper/functions"
	"github.com/4m4rOk/Mr-Proper/mongo"
	"github.com/4m4rOk/Mr-Proper/telegram"
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

	if days < 1 {
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

func AutoIdle(message *tgbotapi.Message) {
	argument := message.CommandArguments()

	var days float64
	days, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		days = 0
	}

	var response tgbotapi.MessageConfig

	if days < 1 && days != 0 {
		response = tgbotapi.NewMessage(message.Chat.ID, message.From.FirstName+".... Nope. Nope. Nope. Nope.")
		response.ReplyToMessageID = message.MessageID
		telegram.Bot.Send(response)
		return
	}

	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var group mongo.Group
	collection.Find(bson.M{"_id": 0}).One(&group)

	if days == 0 {
		if group.AutoIdle != 0 {
			group.AutoIdle = 0
			collection.Update(bson.M{"_id": 0}, &group)
			response = tgbotapi.NewMessage(message.Chat.ID, "No problem, I will stop telling you about inactive users.")
		} else {
			response = tgbotapi.NewMessage(message.Chat.ID, "Let me know how long people can lurk before I notify "+html.EscapeString(message.Chat.Title)+" about them by using <i>/autoidle {{number of days}}</i>.")
		}
	} else {
		if group.AutoIdle == days {
			response = tgbotapi.NewMessage(message.Chat.ID, "You've already told me to notify "+html.EscapeString(message.Chat.Title)+" about peeps which are inactive for more than "+strconv.FormatFloat(days, 'g', -1, 64)+" days...")
		} else {
			group.AutoIdle = days
			collection.Update(bson.M{"_id": 0}, &group)
			response = tgbotapi.NewMessage(message.Chat.ID, "Got it! I will notify "+html.EscapeString(message.Chat.Title)+" about people who lurk for longer than "+strconv.FormatFloat(days, 'g', -1, 64)+" days at a time now!")
		}

	}

	response.ReplyToMessageID = message.MessageID
	response.ParseMode = "HTML"
	telegram.Bot.Send(response)
}

func AutoKick(message *tgbotapi.Message) {
	argument := message.CommandArguments()

	var days float64
	days, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		days = 0
	}

	var response tgbotapi.MessageConfig

	if days < 1 && days != 0 {
		response = tgbotapi.NewMessage(message.Chat.ID, message.From.FirstName+".... Nope. Nope. Nope. Nope.")
		response.ReplyToMessageID = message.MessageID
		telegram.Bot.Send(response)
		return
	}

	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var group mongo.Group
	collection.Find(bson.M{"_id": 0}).One(&group)

	if days == 0 {
		if group.AutoKick != 0 {
			group.AutoKick = 0
			collection.Update(bson.M{"_id": 0}, &group)
			response = tgbotapi.NewMessage(message.Chat.ID, "Oh.. okay. I will stop kicking inactive users by myself!")
		} else {
			response = tgbotapi.NewMessage(message.Chat.ID, "Let me know how long people can lurk before I kick them out of "+html.EscapeString(message.Chat.Title)+" for good! Use <i>/autokick {{number of days}}</i>.")
		}
	} else {
		if group.AutoKick == days {
			response = tgbotapi.NewMessage(message.Chat.ID, "You've already told me to kick peeps which are inactive for more than "+strconv.FormatFloat(days, 'g', -1, 64)+" days...")
		} else {
			group.AutoKick = days
			collection.Update(bson.M{"_id": 0}, &group)
			response = tgbotapi.NewMessage(message.Chat.ID, "Got it! I will kick all them peeps out of "+html.EscapeString(message.Chat.Title)+" if any lurk longer than "+strconv.FormatFloat(days, 'g', -1, 64)+" days at a time now!")
		}

	}

	response.ReplyToMessageID = message.MessageID
	response.ParseMode = "HTML"
	telegram.Bot.Send(response)
}

func GroupsList(message *tgbotapi.Message) {
	groups := functions.GetGroups()

	var response tgbotapi.MessageConfig

	if len(groups) == 0 {
		response = tgbotapi.NewMessage(message.Chat.ID, "I don't have any groups to show. Add me to groups first!")
	} else {
		var text string
		for _, group := range groups {
			groupId, err := strconv.ParseInt(group, 10, 64)
			if err == nil {
				var chatConfig tgbotapi.ChatConfig

				chatConfig.ChatID = groupId
				chat, err := telegram.Bot.GetChat(chatConfig)
				if err == nil {
					var id string
					if chat.UserName == "" {
						id = strconv.FormatInt(chat.ID, 10)
					} else {
						id = "@" + chat.UserName
					}
					text = text + html.EscapeString(chat.Title) + " — " + id + "\n"
				}
			}
		}
		response = tgbotapi.NewMessage(message.Chat.ID, text)
	}
	telegram.Bot.Send(response)
}
