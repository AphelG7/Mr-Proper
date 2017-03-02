package functions

import (
	"github.com/4m4rOk/Mr-Proper/configuration"
	"github.com/4m4rOk/Mr-Proper/mongo"
	"github.com/4m4rOk/Mr-Proper/telegram"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"strings"
)

func GetMember(chat *tgbotapi.Chat, userId int) tgbotapi.ChatMember {
	var chatconfig = chat.ChatConfig()
	var chatconfigwithuser tgbotapi.ChatConfigWithUser

	chatconfigwithuser.ChatID = chatconfig.ChatID
	chatconfigwithuser.SuperGroupUsername = chatconfig.SuperGroupUsername
	chatconfigwithuser.UserID = userId

	member, err := telegram.Bot.GetChatMember(chatconfigwithuser)
	if err != nil {
		log.Fatal(err)
	}
	return member
}

func NewGroup(mchat *tgbotapi.Chat) {
	bot := GetMember(mchat, telegram.Bot.Self.ID)

	if bot.IsAdministrator() != true {
		response := tgbotapi.NewMessage(mchat.ID, "I am no admin. I will not stay.")
		telegram.Bot.Send(response)
		telegram.Bot.LeaveChat(mchat.ChatConfig())
		return
	} else {
		collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(mchat.ID, 10))
		var group mongo.Group
		group.ID = 0
		group.Link = ""
		group.AutoIdle = 0
		group.AutoKick = 0

		collection.Insert(group)

		response := tgbotapi.NewMessage(mchat.ID, "I am the warden of this group. Farewell lurkers and stalkers.")
		telegram.Bot.Send(response)

		if configuration.Config.Mongo.Debug {
			log.Printf("Bot was added to chat %s.", strconv.FormatInt(mchat.ID, 10))
		}
	}
}

func DeleteGroup(mchat *tgbotapi.Chat) {
	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(mchat.ID, 10))

	collection.DropCollection()
	if configuration.Config.Mongo.Debug {
		log.Printf("Removed chat %s.", strconv.FormatInt(mchat.ID, 10))
	}
}

func UpdateUser(muser *tgbotapi.User, mchat *tgbotapi.Chat, mdate int) {
	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(mchat.ID, 10))

	var duser mongo.User
	collection.Find(bson.M{"_id": muser.ID}).One(&duser)

	if duser.ID == muser.ID {
		var updateuser mongo.User
		updateuser.ID = duser.ID
		updateuser.Date = mdate
		updateuser.PermVac = duser.PermVac
		updateuser.TempVac = duser.TempVac

		collection.Update(bson.M{"_id": muser.ID}, &updateuser)
		if configuration.Config.Mongo.Debug {
			log.Printf("Updated user %s in chat %s.", strconv.Itoa(updateuser.ID), strconv.FormatInt(mchat.ID, 10))
		}
	} else {
		var newuser mongo.User
		newuser.ID = muser.ID
		newuser.Date = mdate
		newuser.PermVac = false
		newuser.TempVac = 0

		collection.Insert(newuser)
		if configuration.Config.Mongo.Debug {
			log.Printf("Created user %s in chat %s.", strconv.Itoa(newuser.ID), strconv.FormatInt(mchat.ID, 10))
		}
	}

}

func DeleteUser(muser *tgbotapi.User, mchat *tgbotapi.Chat) {
	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(mchat.ID, 10))

	collection.Remove(bson.M{"_id": muser.ID})
	if configuration.Config.Mongo.Debug {
		log.Printf("Removed user %s in chat %s.", strconv.Itoa(muser.ID), strconv.FormatInt(mchat.ID, 10))
	}
}

func UpdateLink(message *tgbotapi.Message) {
	collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var dgroup mongo.Group
	collection.Find(bson.M{"_id": 0}).One(&dgroup)

	argument := message.CommandArguments()
	var response tgbotapi.MessageConfig

	if strings.HasPrefix(argument, "https://t.me/") || strings.HasPrefix(argument, "https://telegram.me/") {
		var mgroup mongo.Group
		mgroup.ID = 0
		mgroup.Link = argument
		mgroup.AutoIdle = mgroup.AutoIdle
		mgroup.AutoKick = mgroup.AutoKick

		collection.Update(bson.M{"_id": 0}, &mgroup)

		if dgroup.Link == "" {
			response = tgbotapi.NewMessage(message.Chat.ID, "Thanks for letting me know!")

			if configuration.Config.Mongo.Debug {
				log.Printf("Created link %s for chat %s.", argument, strconv.FormatInt(message.Chat.ID, 10))
			}
		} else {
			response = tgbotapi.NewMessage(message.Chat.ID, "Thanks for the new link!")

			if configuration.Config.Mongo.Debug {
				log.Printf("Updated link %s for chat %s.", argument, strconv.FormatInt(message.Chat.ID, 10))
			}
		}
	} else {
		response = tgbotapi.NewMessage(message.Chat.ID, "That does not look like anything to me.")
	}
	response.ReplyToMessageID = message.MessageID
	telegram.Bot.Send(response)
}

func GetGroups() []string {
	collections, err := mongo.Database.DB(configuration.Config.Mongo.Database).CollectionNames()
	if err != nil {
		log.Fatal(err)
	}
	return collections
}
