package db

import (
	"context"
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BlacklistSettings Flood Control struct for user
type BlacklistSettings struct {
	ChatId   int64    `bson:"_id,omitempty" json:"_id,omitempty"`
	Action   string   `bson:"action,omitempty" json:"action,omitempty"`
	Triggers []string `bson:"triggers,omitempty" json:"triggers,omitempty"`
	Reason   string   `bson:"reason,omitempty" json:"reason,omitempty"`
}

// check Chat Blacklists Settings, used to get data before performing any operation
func checkBlacklistSetting(chatID int64) (blSrc *BlacklistSettings) {
	defaultBlacklistSrc := &BlacklistSettings{
		ChatId:   chatID,
		Action:   "none",
		Triggers: make([]string, 0),
		Reason:   "Automated Blacklisted word %s",
	}
	errS := findOne(blacklistsColl, bson.M{"_id": chatID}).Decode(&blSrc)
	if errors.Is(errS, mongo.ErrNoDocuments) {
		blSrc = defaultBlacklistSrc
		err := updateOne(blacklistsColl, bson.M{"_id": chatID}, defaultBlacklistSrc)
		if err != nil {
			log.Errorf("[Database][GetBlacklistSettings]: %v ", err)
		}
	} else if errS != nil {
		log.Errorf("[Database][GetBlacklistSettings]: %v - %d", errS, chatID)
		blSrc = defaultBlacklistSrc
	}
	return blSrc
}

func AddBlacklist(chatId int64, trigger string) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Triggers = append(blSrc.Triggers, strings.ToLower(trigger))
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] AddBlacklist: %v - %d", err, chatId)
	}
}

func RemoveBlacklist(chatId int64, trigger string) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Triggers = removeStrfromStr(blSrc.Triggers, strings.ToLower(trigger))
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", err, chatId)
	}
}

func RemoveAllBlacklist(chatId int64) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Triggers = make([]string, 0)
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] RemoveBlacklist: %v - %d", err, chatId)
	}
}

func SetBlacklistAction(chatId int64, action string) {
	blSrc := checkBlacklistSetting(chatId)
	blSrc.Action = strings.ToLower(action)
	err := updateOne(blacklistsColl, bson.M{"_id": chatId}, blSrc)
	if err != nil {
		log.Errorf("[Database] ChangeBlacklistAction: %v - %d", err, chatId)
	}
}

func GetBlacklistSettings(chatId int64) *BlacklistSettings {
	return checkBlacklistSetting(chatId)
}

func LoadBlacklistsStats() (blacklistTriggers, blacklistChats int64) {
	var blacklistStruct []*BlacklistSettings

	cursor := findAll(blacklistsColl, bson.M{})
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Error(err)
		}
	}(cursor, bgCtx)

	for cursor.Next(bgCtx) {
		var blacklistSetting BlacklistSettings
		if err := cursor.Decode(&blacklistSetting); err != nil {
			log.Error("Failed to decode blacklist setting:", err)
			continue
		}
		blacklistStruct = append(blacklistStruct, &blacklistSetting)
	}

	for _, i := range blacklistStruct {
		lenBl := len(i.Triggers)
		blacklistTriggers += int64(lenBl)
		if lenBl > 0 {
			blacklistChats++
		}
	}

	return
}
