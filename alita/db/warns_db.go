package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type WarnSettings struct {
	ChatId    int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	WarnLimit int    `bson:"warn_limit" json:"warn_limit" default:"3"`
	WarnMode  string `bson:"warn_mode,omitempty" json:"warn_mode,omitempty"`
}

type Warns struct {
	UserId   int64    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	ChatId   int64    `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	NumWarns int      `bson:"num_warns,omitempty" json:"num_warns,omitempty"`
	Reasons  []string `bson:"warns" json:"warns" default:"[]"`
}

func checkWarnSettings(chatID int64) (warnrc *WarnSettings) {
	defaultWarnSettings := &WarnSettings{ChatId: chatID, WarnLimit: 3, WarnMode: "mute"}
	err := findOne(warnSettingsColl, bson.M{"_id": chatID}).Decode(&warnrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		warnrc = defaultWarnSettings
		err := updateOne(warnSettingsColl, bson.M{"_id": chatID}, warnrc)
		if err != nil {
			log.Errorf("[Database] checkWarnSettings: %v", err)
		}
	} else if err != nil {
		log.Errorf("[Database][checkWarnSettings]: %d - %v", chatID, err)
		warnrc = defaultWarnSettings
	}
	return
}

func checkWarns(userId, chatId int64) (warnrc *Warns) {
	defaultWarnSrc := &Warns{UserId: userId, ChatId: chatId, NumWarns: 0, Reasons: make([]string, 0)}
	err := findOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId}).Decode(&warnrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		warnrc = defaultWarnSrc
		err := updateOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId}, warnrc)
		if err != nil {
			log.Errorf("[Database] checkWarns: %v", err)
		}
	} else if err != nil {
		log.Errorf("[Database][checkUserWarns]: %d - %v", userId, err)
		warnrc = defaultWarnSrc
	}
	return
}

func WarnUser(userId, chatId int64, reason string) (int, []string) {
	warnrc := checkWarns(userId, chatId)

	warnrc.NumWarns++ // Increment warns - to Add 1 warn

	// Add reason if it exists
	if reason != "" {
		if len(reason) >= 3001 {
			reason = reason[:3000]
		}
		warnrc.Reasons = append(warnrc.Reasons, reason)
	} else {
		warnrc.Reasons = append(warnrc.Reasons, "No Reason")
	}

	err := updateOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId}, warnrc)
	if err != nil {
		log.Errorf("[Database] WarnUser: %v", err)
	}

	return warnrc.NumWarns, warnrc.Reasons
}

func RemoveWarn(userId, chatId int64) bool {
	removed := false
	warnrc := checkWarns(userId, chatId)

	// only remove if user has warns
	if warnrc.NumWarns > 0 {
		warnrc.NumWarns--                                       // Remove last warn num
		warnrc.Reasons = warnrc.Reasons[:len(warnrc.Reasons)-1] // Remove last warn reason
		removed = true
	}

	// update record in db
	err := updateOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId}, warnrc)
	if err != nil {
		log.Errorf("[Database] RemoveWarn: %v", err)
		return false // force return false to show error
	}

	return removed
}

func ResetUserWarns(userId, chatId int64) (removed bool) {
	removed = true
	err := deleteOne(warnUsersColl, bson.M{"user_id": userId, "chat_id": chatId})
	if err != nil {
		log.Errorf("[Database] ResetUserWarns: %v", err)
		removed = false
	}
	return removed
}

func GetWarns(userId, chatId int64) (int, []string) {
	warnrc := checkWarns(userId, chatId)
	return warnrc.NumWarns, warnrc.Reasons
}

func SetWarnLimit(chatId int64, warnLimit int) {
	warnrc := checkWarnSettings(chatId)
	warnrc.WarnLimit = warnLimit
	err := updateOne(warnSettingsColl, bson.M{"_id": chatId}, warnrc)
	if err != nil {
		log.Errorf("[Database] SetWarnLimit: %v", err)
	}
}

func SetWarnMode(chatId int64, warnMode string) {
	warnrc := checkWarnSettings(chatId)
	warnrc.WarnMode = warnMode
	err := updateOne(warnSettingsColl, bson.M{"_id": chatId}, warnrc)
	if err != nil {
		log.Errorf("[Database] SetWarnMode: %v", err)
	}
}

func GetWarnSetting(chatId int64) *WarnSettings {
	return checkWarnSettings(chatId)
}

func GetAllChatWarns(chatId int64) int {
	length, _ := countDocs(warnUsersColl, bson.M{"chat_id": chatId})
	return int(length)
}

func ResetAllChatWarns(chatId int64) bool {
	err := deleteMany(warnUsersColl, bson.M{"chat_id": chatId})
	if err != nil {
		log.Errorf("[Database] ResetAllChatWarns: %v", err)
		return false
	}
	return true
}
