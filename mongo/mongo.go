package mongo

import (
	"gopkg.in/mgo.v2"
	"github.com/4m4rOk/Mr-Proper/configuration"
)

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

var Database, _ = mgo.Dial(configuration.Config.Mongo.Url)