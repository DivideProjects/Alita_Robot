package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// AdminSettings Flood Settings struct for chat
type AdminSettings struct {
	ChatId    int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	AnonAdmin bool  `bson:"anon_admin" json:"anon_admin"`
}

// GetAdminSettings Get admin settings for a chat
func GetAdminSettings(chatID int64) *AdminSettings {
	return checkAdminSetting(chatID)
}

// check Chat Admin Settings, used to get data before performing any operation
func checkAdminSetting(chatID int64) (adminSrc *AdminSettings) {
	defaultAdminSrc := &AdminSettings{ChatId: chatID, AnonAdmin: false}

	err := findOne(adminSettingsColl, bson.M{"_id": chatID}).Decode(&adminSrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		adminSrc = defaultAdminSrc
		err := updateOne(adminSettingsColl, bson.M{"_id": chatID}, defaultAdminSrc)
		if err != nil {
			log.Errorf("[Database][checkAdminSetting]: %v ", err)
		}
	} else if err != nil {
		adminSrc = defaultAdminSrc
		log.Errorf("[Database][checkAdminSetting]: %v ", err)
	}
	return adminSrc
}

// SetAnonAdminMode Set anon admin mode for a chat
func SetAnonAdminMode(chatID int64, val bool) {
	adminSrc := checkAdminSetting(chatID)
	adminSrc.AnonAdmin = val

	err := updateOne(adminSettingsColl, bson.M{"_id": chatID}, adminSrc)
	if err != nil {
		log.Errorf("[Database] SetAnonAdminMode: %v - %d", err, chatID)
	}
}
