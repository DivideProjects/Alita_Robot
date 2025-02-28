package db

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	UserId   int64  `bson:"_id,omitempty" json:"_id,omitempty"`
	UserName string `bson:"username" json:"username" default:"nil"`
	Name     string `bson:"name" json:"name" default:"nil"`
	Language string `bson:"language" json:"language" default:"en"`
}

func EnsureBotInDb(b *gotgbot.Bot) {
	usersUpdate := &User{UserId: b.Id, UserName: b.Username, Name: b.FirstName}
	err := updateOne(userColl, bson.M{"_id": b.Id}, usersUpdate)
	if err != nil {
		log.Errorf("[Database] EnsureBotInDb: %v", err)
	}
	log.Infof("[Database] Bot Updated in Database!")
}

func checkUserInfo(userId int64) (userc *User) {
	defaultUser := &User{UserId: userId}
	errS := findOne(userColl, bson.M{"_id": userId}).Decode(&userc)
	if errors.Is(errS, mongo.ErrNoDocuments) {
		userc = nil
	} else if errS != nil {
		log.Errorf("[Database] checkUserInfo: %v - %d", errS, userId)
		userc = defaultUser
	}
	return userc
}

func UpdateUser(userId int64, username, name string) {
	userc := checkUserInfo(userId)

	if userc != nil {
		if userc.Name == name && userc.UserName == username {
			return
		}
		userc.Name = name
		userc.UserName = username
	} else {
		userc = &User{
			UserId:   userId,
			UserName: username,
			Name:     name,
		}
	}

	err2 := updateOne(userColl, bson.M{"_id": userId}, userc)
	if err2 != nil {
		log.Errorf("[Database] UpdateUser: %v - %d", err2, userId)
		return
	}
	log.Infof("[Database] UpdateUser: %d", userId)
}

func GetUserIdByUserName(username string) int64 {
	var guids *User
	err := findOne(userColl, bson.M{"username": username}).Decode(&guids)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0
	} else if err != nil {
		log.Errorf("[Database] GetUserIdByUserName: %v - %d", err, guids.UserId)
		return 0
	}
	log.Infof("[Database] GetUserIdByUserName: %d", guids.UserId)
	return guids.UserId
}

func GetUserInfoById(userId int64) (username, name string, found bool) {
	user := checkUserInfo(userId)
	if user != nil {
		username = user.UserName
		name = user.Name
		found = true
		log.Debugf("%+v", user)
	}
	return
}

func LoadUsersStats() (count int64) {
	count, err := countDocs(userColl, bson.M{})
	if err != nil {
		log.Errorf("[Database] loadStats: %v", err)
		return
	}
	return
}
