package main

import (
	"github.com/BurntSushi/toml"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html"
	"log"
	"os"
	"strconv"
	"strings"
)

var Config = readConfig("mrproper.config")
var Mongo, _ = mgo.Dial(Config.Mongo.Url)
var Bot, _ = tgbotapi.NewBotAPI(Config.Telegram.Token)

func main() {
	defer Mongo.Close()

	Bot.Debug = Config.Telegram.Debug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := Bot.GetUpdatesChan(u)
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

type User struct {
	ID      int  `bson:"_id"`
	Date    int  `bson:"Date"`
	PermVac bool `bson:"PermVac"`
	TempVac int  `bson:"TempVac"`
}

type Group struct {
	ID   int    `bson:"_id"`
	Link string `bson:"Link"`
}

type configInfo struct {
	Telegram telegramConfig
	Mongo    mongoConfig
}

type telegramConfig struct {
	Token string
	Debug bool
}

type mongoConfig struct {
	Url      string
	Database string
	Debug    bool
}

func readConfig(configfile string) configInfo {
	_, err := os.Open(configfile)
	if err != nil {
		log.Fatal(err)
	}

	var config configInfo
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	return config
}

func handleGroupCommand(message *tgbotapi.Message) {
	user := getMember(message.Chat, message.From.ID)

	if message.Command() == "link" {
		linkCommand(message)
	} else if message.Command() == "id" {
		idCommand(message)
	}

	if user.IsAdministrator() || user.IsCreator() {
		if message.Command() == "idle" {
			idleCommand(message)
		} else if message.Command() == "kick" {
			if message.Chat.IsSuperGroup() == false {
				response := tgbotapi.NewMessage(message.Chat.ID, "I can only kick in supergroups, sorry!")
				Bot.Send(response)
			} else {
				kickCommand(message)
			}
		}
	}
}

func handleGroupMessage(message *tgbotapi.Message) {
	if message.NewChatMember != nil {
		if message.NewChatMember.ID == Bot.Self.ID {
			newGroup(message.Chat)
			return
		} else {
			updateUser(message.NewChatMember, message.Chat, message.Date)
			return
		}
	}
	if message.LeftChatMember != nil {
		if message.LeftChatMember.ID == Bot.Self.ID {
			deleteGroup(message.Chat)
			return
		} else {
			deleteUser(message.LeftChatMember, message.Chat)
			return
		}
	}
	updateUser(message.From, message.Chat, message.Date)
}

func handlePrivate(message *tgbotapi.Message) {
	response := tgbotapi.NewMessage(message.Chat.ID, "Hello dear creator. Make me admin in a supergroup of yours and I will show you how to rule properly.")
	Bot.Send(response)
}

func getMember(chat *tgbotapi.Chat, userId int) tgbotapi.ChatMember {
	var chatconfig = chat.ChatConfig()
	var chatconfigwithuser tgbotapi.ChatConfigWithUser

	chatconfigwithuser.ChatID = chatconfig.ChatID
	chatconfigwithuser.SuperGroupUsername = chatconfig.SuperGroupUsername
	chatconfigwithuser.UserID = userId

	member, err := Bot.GetChatMember(chatconfigwithuser)
	if err != nil {
		log.Fatal(err)
	}
	return member
}

func newGroup(mchat *tgbotapi.Chat) {
	bot := getMember(mchat, Bot.Self.ID)

	if bot.IsAdministrator() != true {
		response := tgbotapi.NewMessage(mchat.ID, "I am no admin. I will not stay.")
		Bot.Send(response)
		Bot.LeaveChat(mchat.ChatConfig())
		return
	} else {
		response := tgbotapi.NewMessage(mchat.ID, "I am the warden of this group. Farewell lurkers and stalkers.")
		Bot.Send(response)

		if Config.Mongo.Debug {
			log.Printf("Bot was added to chat %s.", strconv.FormatInt(mchat.ID, 10))
		}
	}
}

func deleteGroup(mchat *tgbotapi.Chat) {
	collection := Mongo.DB(Config.Mongo.Database).C(strconv.FormatInt(mchat.ID, 10))

	collection.DropCollection()
	if Config.Mongo.Debug {
		log.Printf("Removed chat %s.", strconv.FormatInt(mchat.ID, 10))
	}
}

func updateUser(muser *tgbotapi.User, mchat *tgbotapi.Chat, mdate int) {
	collection := Mongo.DB(Config.Mongo.Database).C(strconv.FormatInt(mchat.ID, 10))

	var duser User
	collection.Find(bson.M{"_id": muser.ID}).One(&duser)

	if duser.ID == muser.ID {
		var updateuser User
		updateuser.ID = duser.ID
		updateuser.Date = mdate
		updateuser.PermVac = duser.PermVac
		updateuser.TempVac = duser.TempVac

		collection.Update(bson.M{"_id": muser.ID}, &updateuser)
		if Config.Mongo.Debug {
			log.Printf("Updated user %s in chat %s.", strconv.Itoa(updateuser.ID), strconv.FormatInt(mchat.ID, 10))
		}
	} else {
		var newuser User
		newuser.ID = muser.ID
		newuser.Date = mdate
		newuser.PermVac = false
		newuser.TempVac = 0

		collection.Insert(newuser)
		if Config.Mongo.Debug {
			log.Printf("Created user %s in chat %s.", strconv.Itoa(newuser.ID), strconv.FormatInt(mchat.ID, 10))
		}
	}

}

func deleteUser(muser *tgbotapi.User, mchat *tgbotapi.Chat) {
	collection := Mongo.DB(Config.Mongo.Database).C(strconv.FormatInt(mchat.ID, 10))

	collection.Remove(bson.M{"_id": muser.ID})
	if Config.Mongo.Debug {
		log.Printf("Removed user %s in chat %s.", strconv.Itoa(muser.ID), strconv.FormatInt(mchat.ID, 10))
	}
}

func updateLink(message *tgbotapi.Message) {
	collection := Mongo.DB(Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var dgroup Group
	collection.Find(bson.M{"_id": 0}).One(&dgroup)

	argument := message.CommandArguments()
	var response tgbotapi.MessageConfig

	log.Printf(dgroup.Link)

	if strings.HasPrefix(argument, "https://t.me/") || strings.HasPrefix(argument, "https://telegram.me/") {
		if dgroup.Link == "" {
			var mgroup Group
			mgroup.ID = 0
			mgroup.Link = argument

			collection.Insert(mgroup)

			response = tgbotapi.NewMessage(message.Chat.ID, "Thanks for letting me know!")

			if Config.Mongo.Debug {
				log.Printf("Created link %s for chat %s.", argument, strconv.FormatInt(message.Chat.ID, 10))
			}
		} else {
			var mgroup Group
			mgroup.ID = 0
			mgroup.Link = argument

			collection.Update(bson.M{"_id": 0}, &mgroup)

			response = tgbotapi.NewMessage(message.Chat.ID, "Thanks for the new link!")

			if Config.Mongo.Debug {
				log.Printf("Updated link %s for chat %s.", argument, strconv.FormatInt(message.Chat.ID, 10))
			}
		}
	} else {
		response = tgbotapi.NewMessage(message.Chat.ID, "That does not look like anything to me.")
	}
	response.ReplyToMessageID = message.MessageID
	Bot.Send(response)
}

func idCommand(message *tgbotapi.Message) {
	var response tgbotapi.MessageConfig

	if message.ReplyToMessage == nil {
		response = tgbotapi.NewMessage(message.Chat.ID, "<b>"+html.EscapeString(message.From.FirstName)+"'s ID:</b> "+strconv.Itoa(message.From.ID)+"\n<b>"+html.EscapeString(message.Chat.Title)+"'s ID:</b> "+strconv.FormatInt(message.Chat.ID, 10))
	} else {
		response = tgbotapi.NewMessage(message.Chat.ID, "<b>"+html.EscapeString(message.ReplyToMessage.From.FirstName)+"'s ID:</b> "+strconv.Itoa(message.ReplyToMessage.From.ID)+"\n<b>Message ID:</b> "+strconv.Itoa(message.ReplyToMessage.MessageID))
	}

	response.ReplyToMessageID = message.MessageID
	response.ParseMode = "HTML"
	Bot.Send(response)
}

func linkCommand(message *tgbotapi.Message) {
	collection := Mongo.DB(Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	argument := message.CommandArguments()

	if argument == "" {
		var group Group
		collection.Find(bson.M{"_id": 0}).One(&group)

		var response tgbotapi.MessageConfig
		if group.Link != "" {
			response = tgbotapi.NewMessage(message.Chat.ID, html.EscapeString(message.From.FirstName)+", the link is:\n"+group.Link)
		} else {
			response = tgbotapi.NewMessage(message.Chat.ID, "I don't know the link. Sorry!")
		}
		response.ReplyToMessageID = message.MessageID
		Bot.Send(response)
	} else {
		user := getMember(message.Chat, message.From.ID)
		if user.IsAdministrator() || user.IsCreator() {
			updateLink(message)
		}
	}
}

func idleCommand(message *tgbotapi.Message) {
	firstResponse := tgbotapi.NewMessage(message.Chat.ID, "Hang on... Let me see who has been naughty!")
	Bot.Send(firstResponse)

	argument := message.CommandArguments()

	var days float64
	days, err := strconv.ParseFloat(argument, 64)
	if err != nil {
		days = 7.0
	}

	collection := Mongo.DB(Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var dusers []User
	collection.Find(bson.M{"Date": bson.M{"$lt": (message.Date - int(days*86400))}}).All(&dusers)

	var text string
	for _, duser := range dusers {
		member := getMember(message.Chat, duser.ID)

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
	Bot.Send(secondResponse)
}

func kickCommand(message *tgbotapi.Message) {
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
		Bot.Send(response)
		return
	}

	collection := Mongo.DB(Config.Mongo.Database).C(strconv.FormatInt(message.Chat.ID, 10))

	var dusers []User
	collection.Find(bson.M{"Date": bson.M{"$lt": (message.Date - int(days*86400))}}).All(&dusers)

	for _, duser := range dusers {
		member := getMember(message.Chat, duser.ID)

		if member.IsCreator() {
			continue
		}

		var chatConfig = message.Chat.ChatConfig()
		var chatMemberConfig tgbotapi.ChatMemberConfig

		chatMemberConfig.ChatID = chatConfig.ChatID
		chatMemberConfig.SuperGroupUsername = chatConfig.SuperGroupUsername
		chatMemberConfig.UserID = duser.ID

		Bot.KickChatMember(chatMemberConfig)
		collection.Remove(bson.M{"_id": duser.ID})
		Bot.UnbanChatMember(chatMemberConfig)

	}

	if len(dusers) == 0 {
		response = tgbotapi.NewMessage(message.Chat.ID, "No one was inactive for more than "+strconv.FormatFloat(days, 'g', -1, 64)+" days. Happy?")
	} else {
		response = tgbotapi.NewMessage(message.Chat.ID, "Happy?")
	}

	response.ReplyToMessageID = message.MessageID
	Bot.Send(response)
}
