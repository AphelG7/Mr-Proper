package main

import (
	"github.com/4m4rOk/Mr-Proper/configuration"
	"github.com/4m4rOk/Mr-Proper/functions"
	"github.com/4m4rOk/Mr-Proper/mongo"
	"github.com/4m4rOk/Mr-Proper/telegram"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/robfig/cron.v2"
	"strconv"
	//"log"
)

func main() {
	defer mongo.Database.Close()

	c := cron.New()
	c.AddFunc("@every 10s", scheduledJobs)
	c.Start()

	select {}
}

func scheduledJobs() {
	groups := functions.GetGroups()

	for _, groupId := range groups {
		collection := mongo.Database.DB(configuration.Config.Mongo.Database).C(groupId)
		var group mongo.Group
		collection.Find(bson.M{"_id": 0}).One(&group)
		text := group.Link + "\n"
		chatId, err := strconv.ParseInt(groupId, 10, 64)
		if err == nil {
			response := tgbotapi.NewMessage(chatId, text)
			telegram.Bot.Send(response)
		}
	}
}
