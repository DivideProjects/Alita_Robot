package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Channel struct {
	ChannelId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	ChannelName string `bson:"channel_name" json:"channel_name" default:"nil"`
	Username    string `bson:"username" json:"username" default:"nil"`
}

func GetChannelSettings(channelId int64) (channelSrc *Channel) {
	err := findOne(channelColl, bson.M{"_id": channelId}).Decode(&channelSrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		channelSrc = nil
	} else if err != nil {
		log.Errorf("[Database] getChannelSettings: %v - %d ", err, channelId)
		return
	}
	return
}

func UpdateChannel(channelId int64, channelName, username string) {
	channelSrc := GetChannelSettings(channelId)

	if channelSrc != nil {
		if channelSrc.ChannelName == channelName && channelSrc.Username == username {
			return
		}
		channelSrc.ChannelName = channelName
		channelSrc.Username = username
	} else {
		channelSrc = &Channel{
			ChannelId:   channelId,
			ChannelName: channelName,
			Username:    username,
		}
	}

	err2 := updateOne(channelColl, bson.M{"_id": channelId}, channelSrc)
	if err2 != nil {
		log.Errorf("[Database] UpdateChannel: %v - %d (%s)", err2, channelId, username)
		return
	}
	log.Infof("[Database] UpdateChannel: %s", channelName)
}

func GetChannelIdByUserName(username string) int64 {
	var cuids *Channel
	err := findOne(channelColl, bson.M{"username": username}).Decode(&cuids)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0
	} else if err != nil {
		log.Errorf("[Database] GetChannelByUserName: %v - %d", err, cuids.ChannelId)
		return 0
	}
	log.Infof("[Database] GetChannelByUserName: %d", cuids.ChannelId)
	return cuids.ChannelId
}

func GetChannelInfoById(userId int64) (username, name string, found bool) {
	channel := GetChannelSettings(userId)
	if channel != nil {
		username = channel.Username
		name = channel.ChannelName
		found = true
	}
	return
}

func LoadChannelStats() (count int64) {
	count, err := countDocs(channelColl, bson.M{})
	if err != nil {
		log.Errorf("[Database] loadChannelStats: %v", err)
		return 0
	}
	return
}
