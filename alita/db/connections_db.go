package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Connections struct {
	UserId    int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	ChatId    int64 `bson:"chat_id,omitempty" json:"chat_id,omitempty"`
	Connected bool  `bson:"connected" json:"connected" default:"false"`
}

type ConnectionSettings struct {
	ChatId       int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	AllowConnect bool  `bson:"can_connect" json:"can_connect" default:"false"`
}

func ToggleAllowConnect(chatID int64, pref bool) {
	connectionSrc := GetChatConnectionSetting(chatID)
	connectionSrc.AllowConnect = pref
	err := updateOne(connectionSettingsColl, bson.M{"_id": chatID}, connectionSrc)
	if err != nil {
		log.Errorf("[Database] ToggleAllowConnect: %d - %v", chatID, err)
	}
}

func GetChatConnectionSetting(chatID int64) (connectionSrc *ConnectionSettings) {
	defaultConnectionSrc := &ConnectionSettings{ChatId: chatID, AllowConnect: false}
	errF := findOne(connectionSettingsColl, bson.M{"_id": chatID}).Decode(&connectionSrc)
	if errors.Is(errF, mongo.ErrNoDocuments) {
		connectionSrc = defaultConnectionSrc
		err := updateOne(connectionSettingsColl, bson.M{"_id": chatID}, connectionSrc)
		if err != nil {
			log.Errorf("[Database] GetChatConnectionSetting: %d - %v", chatID, err)
		}
	} else if errF != nil {
		connectionSrc = defaultConnectionSrc
		log.Errorf("[Database] GetChatSetting: %d - %v", chatID, errF)
	}
	return connectionSrc
}

func getUserConnectionSetting(userID int64) (connectionSrc *Connections) {
	defaultConnectionSrc := &Connections{UserId: userID, Connected: false}
	errF := findOne(connectionColl, bson.M{"_id": userID}).Decode(&connectionSrc)
	if errors.Is(errF, mongo.ErrNoDocuments) {
		connectionSrc = defaultConnectionSrc
		err := updateOne(connectionColl, bson.M{"_id": userID}, connectionSrc)
		if err != nil {
			log.Errorf("[Database] GetChatConnectionSetting: %d - %v", userID, err)
		}
	} else if errF != nil {
		connectionSrc = defaultConnectionSrc
		log.Errorf("[Database] GetUserSetting: %d - %v", userID, errF)
	}

	return connectionSrc
}

func Connection(UserID int64) *Connections {
	return getUserConnectionSetting(UserID)
}

func ConnectId(UserID, chatID int64) {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = true
	connectionUpdate.ChatId = chatID
	err := updateOne(connectionColl, bson.M{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] ConnectId: %v - %d", err, chatID)
	}
}

func DisconnectId(UserID int64) {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = false
	err := updateOne(connectionColl, bson.M{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] DisconnectId: %v - %d", err, UserID)
	}
}

func ReconnectId(UserID int64) int64 {
	connectionUpdate := Connection(UserID)
	connectionUpdate.Connected = true
	err := updateOne(connectionColl, bson.M{"_id": UserID}, connectionUpdate)
	if err != nil {
		log.Errorf("[Database] ReconnectId: %v - %d", err, UserID)
		return 0
	}
	return connectionUpdate.ChatId
}

func LoadConnectionStats() (connectedUsers, connectedChats int64) {
	connectedChats, err := countDocs(
		connectionSettingsColl,
		bson.M{"can_connect": true},
	)
	if err != nil {
		log.Error(err)
		return
	}

	connectedUsers, err = countDocs(
		connectionColl,
		bson.M{"connected": true},
	)
	if err != nil {
		log.Error(err)
		return
	}

	return
}
