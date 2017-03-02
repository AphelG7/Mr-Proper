package mongo

import (
	"github.com/4m4rOk/Mr-Proper/configuration"
	"gopkg.in/mgo.v2"
)

type User struct {
	ID      int  `bson:"_id"`
	Date    int  `bson:"Date"`
	PermVac bool `bson:"PermVac"`
	TempVac int  `bson:"TempVac"`
}

type Group struct {
	ID       int     `bson:"_id"`
	Link     string  `bson:"Link"`
	AutoIdle float64 `bson:"AutoIdle"`
	AutoKick float64 `bson:"AutoKick"`
}

var Database, _ = mgo.Dial(configuration.Config.Mongo.Url)
