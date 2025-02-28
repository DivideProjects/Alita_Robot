package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"
)

type DisableCommand struct {
	ChatID       int64    `bson:"_id,omitempty" json:"_id,omitempty"`
	Commands     []string `bson:"commands" json:"commands" default:"none"`
	ShouldDelete bool     `bson:"should_delete" json:"should_delete" default:"false"`
}

// check Chat Disable Settings, used to get data before performing any operation
func checkDisableSettings(chatID int64) (disSrc *DisableCommand) {
	defaultDisrc := &DisableCommand{ChatID: chatID, Commands: make([]string, 0), ShouldDelete: false}
	errS := findOne(disableColl, bson.M{"_id": chatID}).Decode(&disSrc)
	if errors.Is(errS, mongo.ErrNoDocuments) {
		disSrc = defaultDisrc
		err := updateOne(disableColl, bson.M{"_id": chatID}, defaultDisrc)
		if err != nil {
			log.Errorf("[Database] checkDisableSettings: %d - %v", chatID, err)
		}
	} else if errS != nil {
		log.Errorf("[Database][checkDisableSettings]: %v", errS)
		disSrc = defaultDisrc
	}
	return disSrc
}

// DisableCMD Disable CMD in chat
func DisableCMD(chatID int64, cmd string) {
	disableCmd := checkDisableSettings(chatID)
	disableCmd.Commands = append(disableCmd.Commands, cmd)
	err := updateOne(disableColl, bson.M{"_id": chatID}, disableCmd)
	if err != nil {
		log.Errorf("[Database][DisableCMD]: %v", err)
	}
}

// EnableCMD Enable CMD in chat
func EnableCMD(chatID int64, cmd string) {
	disableCmd := checkDisableSettings(chatID)
	disableCmd.Commands = removeStrfromStr(disableCmd.Commands, cmd)
	err := updateOne(disableColl, bson.M{"_id": chatID}, disableCmd)
	if err != nil {
		log.Errorf("[Database][EnableCMD]: %v", err)
	}
}

// GetChatDisabledCMDs Get disabled comands of chat
func GetChatDisabledCMDs(chatId int64) []string {
	return checkDisableSettings(chatId).Commands
}

// IsCommandDisabled Check if command is disabled or not
func IsCommandDisabled(chatId int64, cmd string) bool {
	return string_handling.FindInStringSlice(GetChatDisabledCMDs(chatId), cmd)
}

// ToggleDel Toggle Command Deleting
func ToggleDel(chatId int64, pref bool) {
	disableCmd := checkDisableSettings(chatId)
	disableCmd.ShouldDelete = pref
	err := updateOne(disableColl, bson.M{"_id": chatId}, disableCmd)
	if err != nil {
		log.Error(err)
	}
}

// ShouldDel Check if cmd del is enabled or not
func ShouldDel(chatId int64) bool {
	disableCmd := checkDisableSettings(chatId)
	return disableCmd.ShouldDelete
}

// remove a string element from an string slice
func removeStrfromStr(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func LoadDisableStats() (disabledCmds, disableEnabledChats int64) {
	var disbaledStruct []*DisableCommand

	cursor := findAll(disableColl, bson.M{})
	defer cursor.Close(bgCtx)
	cursor.All(bgCtx, &disbaledStruct)

	for _, disrc := range disbaledStruct {
		disLn := int64(len(disrc.Commands))
		disabledCmds += disLn
		if disLn > 0 {
			disableEnabledChats++
		}
	}

	return
}
