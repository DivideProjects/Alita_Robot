package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// default mode is 'mute'
const defaultFloodsettingsMode string = "mute"

// FloodSettings Flood Settings struct for chat
type FloodSettings struct {
	ChatId                 int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	Limit                  int    `bson:"limit" json:"limit"`
	Mode                   string `bson:"mode,omitempty" json:"mode,omitempty"`
	DeleteAntifloodMessage bool   `bson:"del_msg" json:"del_msg"`
}

// GetFlood Get flood settings for a chat
func GetFlood(chatID int64) *FloodSettings {
	return checkFloodSetting(chatID)
}

// check Chat Flood Settings, used to get data before performing any operation
func checkFloodSetting(chatID int64) (floodSrc *FloodSettings) {
	defaultFloodSrc := &FloodSettings{ChatId: chatID, Limit: 0, Mode: defaultFloodsettingsMode}

	err := findOne(antifloodSettingsColl, bson.M{"_id": chatID}).Decode(&floodSrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		floodSrc = defaultFloodSrc
		err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, defaultFloodSrc)
		if err != nil {
			log.Errorf("[Database][checkFloodSetting]: %v ", err)
		}
	} else if err != nil {
		floodSrc = defaultFloodSrc
		log.Errorf("[Database][checkGreetingSettings]: %v ", err)
	}
	return floodSrc
}

// SetFlood set Flood Setting for a Chat
func SetFlood(chatID int64, limit int) {
	floodSrc := checkFloodSetting(chatID)

	if floodSrc.Mode == "" {
		floodSrc = &FloodSettings{ChatId: chatID, Limit: limit, Mode: defaultFloodsettingsMode}
	} else {
		floodSrc.Limit = limit // update floodSrc.limit
	}

	// update the value in db
	err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFlood: %v - %d", err, chatID)
	}
}

// SetFloodMode Set flood mode for a chat
func SetFloodMode(chatID int64, mode string) {
	floodSrc := checkFloodSetting(chatID)
	floodSrc.Mode = mode

	err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFloodMode: %v - %d", err, chatID)
	}
}

// SetFloodMsgDel Set flood mode for a chat
func SetFloodMsgDel(chatID int64, val bool) {
	floodSrc := checkFloodSetting(chatID)
	floodSrc.DeleteAntifloodMessage = val

	err := updateOne(antifloodSettingsColl, bson.M{"_id": chatID}, floodSrc)
	if err != nil {
		log.Errorf("[Database] SetFloodMsgDel: %v - %d", err, chatID)
	}
}

func LoadAntifloodStats() (antiCount int64) {
	totalCount, err := countDocs(antifloodSettingsColl, bson.M{})
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
	}
	noAntiCount, err := countDocs(antifloodSettingsColl, bson.M{"limit": 0})
	if err != nil {
		log.Errorf("[Database] LoadAntifloodStats: %v", err)
	}

	antiCount = totalCount - noAntiCount //  gives chats which have enabled anti flood

	return
}
