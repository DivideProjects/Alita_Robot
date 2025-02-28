package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatReportSettings struct {
	ChatId      int64   `bson:"_id,omitempty" json:"_id,omitempty"`
	Status      bool    `bson:"status,omitempty" json:"status,omitempty"`
	BlockedList []int64 `bson:"blocked_list,omitempty" json:"blocked_list,omitempty"`
}

type UserReportSettings struct {
	UserId int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	Status bool  `bson:"status,omitempty" json:"status,omitempty"`
}

func GetChatReportSettings(chatID int64) (reportsrc *ChatReportSettings) {
	defReportSettings := &ChatReportSettings{
		ChatId:      chatID,
		Status:      true,
		BlockedList: make([]int64, 0),
	}

	err := findOne(reportChatColl, bson.M{"_id": chatID}).Decode(&reportsrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		reportsrc = defReportSettings
		err := updateOne(reportChatColl, bson.M{"_id": chatID}, reportsrc)
		if err != nil {
			log.Error(err)
		}
	} else if err != nil {
		reportsrc = defReportSettings
		log.Error(err)
	}
	return
}

func SetChatReportStatus(chatID int64, pref bool) {
	reportsUpdate := GetChatReportSettings(chatID)
	reportsUpdate.Status = pref
	err := updateOne(reportChatColl, bson.M{"_id": chatID}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
}

func BlockReportUser(chatId, userId int64) {
	reportsUpdate := GetChatReportSettings(chatId)

	if string_handling.FindInInt64Slice(reportsUpdate.BlockedList, userId) {
		return
	}

	reportsUpdate.BlockedList = append(reportsUpdate.BlockedList, userId)
	err := updateOne(reportChatColl, bson.M{"_id": chatId}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
}

func UnblockReportUser(chatId, userId int64) {
	reportsUpdate := GetChatReportSettings(chatId)

	if !string_handling.FindInInt64Slice(reportsUpdate.BlockedList, userId) {
		return
	}

	reportsUpdate.BlockedList = string_handling.RemoveFromInt64Slice(reportsUpdate.BlockedList, userId)
	err := updateOne(reportChatColl, bson.M{"_id": chatId}, reportsUpdate)
	if err != nil {
		log.Error(err)
	}
}

/* user settings below */

func GetUserReportSettings(userId int64) (reportsrc *UserReportSettings) {
	defReportSettings := &UserReportSettings{
		UserId: userId,
		Status: true,
	}

	err := findOne(reportUserColl, bson.M{"_id": userId}).Decode(&reportsrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		reportsrc = defReportSettings
		err := updateOne(reportUserColl, bson.M{"_id": userId}, reportsrc)
		if err != nil {
			log.Error(err)
		}
	} else if err != nil {
		reportsrc = defReportSettings
		log.Error(err)
	}

	return
}

func SetUserReportSettings(chatID int64, pref bool) {
	reportsUpdate := &ChatReportSettings{
		ChatId: chatID,
		Status: pref,
	}
	err := updateOne(reportUserColl, bson.M{"_id": chatID}, reportsUpdate)
	if err != nil {
		log.Error(chatID)
	}
}

func LoadReportStats() (uRCount, gRCount int64) {
	uRCount, acErr := countDocs(
		reportUserColl,
		bson.M{"status": true},
	)
	if acErr != nil {
		log.Errorf("[Database] loadStats: %v", acErr)
	}
	gRCount, clErr := countDocs(
		reportChatColl,
		bson.M{"status": true},
	)
	if clErr != nil {
		log.Errorf("[Database] loadStats: %v", clErr)
	}
	return
}
