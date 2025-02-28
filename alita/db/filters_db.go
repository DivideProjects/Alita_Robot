package db

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

type ChatFilters struct {
	ChatId      int64    `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	KeyWord     string   `bson:"keyword,omitempty" json:"keyword,omitempty"`
	FilterReply string   `bson:"filter_reply,omitempty" json:"filter_reply,omitempty"`
	MsgType     int      `bson:"msgtype,omitempty" json:"msgtype,omitempty"`
	FileID      string   `bson:"fileid,omitempty" json:"fileid,omitempty"`
	NoNotif     bool     `bson:"nonotif,omitempty" json:"nonotif,omitempty"`
	Buttons     []Button `bson:"filter_buttons,omitempty" json:"filter_buttons,omitempty"`
}

func GetFilter(chatID int64, keyword string) (filtSrc *ChatFilters) {
	err := findOne(filterColl, bson.M{"chat_id": chatID, "keyword": keyword}).Decode(&filtSrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		filtSrc = &ChatFilters{}
	} else if err != nil {
		log.Errorf("[Database] GetFilter: %v - %d", err, chatID)
	}
	return
}

//goland:noinspection GoUnusedExportedFunction
func GetAllFilters(chatID int64) (allFilters []*ChatFilters) {
	cursor := findAll(filterColl, bson.M{"chat_id": chatID})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &allFilters)
	return
}

func GetFiltersList(chatID int64) (allFilterWords []string) {
	var results []*ChatFilters
	cursor := findAll(filterColl, bson.M{"chat_id": chatID})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &results)
	for _, j := range results {
		allFilterWords = append(allFilterWords, j.KeyWord)
	}
	return
}

func DoesFilterExists(chatId int64, keyword string) bool {
	return string_handling.FindInStringSlice(GetFiltersList(chatId), strings.ToLower(keyword))
}

func AddFilter(chatID int64, keyWord, replyText, fileID string, buttons []Button, filtType int) {
	if string_handling.FindInStringSlice(GetFiltersList(chatID), keyWord) {
		return
	}

	// add the filter
	newFilter := ChatFilters{
		ChatId:      chatID,
		KeyWord:     keyWord,
		FilterReply: replyText,
		MsgType:     filtType,
		FileID:      fileID,
		Buttons:     buttons,
	}

	err := updateOne(filterColl, bson.M{"chat_id": chatID, "keyword": keyWord}, newFilter)
	if err != nil {
		log.Errorf("[Database][AddFilter]: %d - %v", chatID, err)
		return
	}
}

func RemoveFilter(chatID int64, keyWord string) {
	if !string_handling.FindInStringSlice(GetFiltersList(chatID), keyWord) {
		return
	}

	err := deleteOne(filterColl, bson.M{"chat_id": chatID, "keyword": keyWord})
	if err != nil {
		log.Errorf("[Database][RemoveFilter]: %d - %v", chatID, err)
		return
	}
}

func RemoveAllFilters(chatID int64) {
	err := deleteMany(filterColl, bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][RemoveAllFilters]: %d - %v", chatID, err)
	}
}

func CountFilters(chatID int64) (filtersNum int64) {
	filtersNum, err := countDocs(filterColl, bson.M{"chat_id": chatID})
	if err != nil {
		log.Errorf("[Database][CountFilters]: %d - %v", chatID, err)
	}
	return
}

func LoadFilterStats() (filtersNum, filtersUsingChats int64) {
	var filterStruct []*ChatFilters
	filtersMap := make(map[int64][]ChatFilters)

	cursor := findAll(filterColl, bson.M{})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &filterStruct)

	for _, filterC := range filterStruct {
		filtersNum++ // count number of filters
		filtersMap[filterC.ChatId] = append(filtersMap[filterC.ChatId], *filterC)
	}

	filtersUsingChats = int64(len(filtersMap))

	return
}
