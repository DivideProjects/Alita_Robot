package db

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Button struct {
	Name     string `bson:"name,omitempty" json:"name,omitempty"`
	Url      string `bson:"url,omitempty" json:"url,omitempty"`
	SameLine bool   `bson:"btn_sameline" json:"btn_sameline" default:"false"`
}

const (
	// TEXT types of senders
	TEXT      int = 1
	STICKER   int = 2
	DOCUMENT  int = 3
	PHOTO     int = 4
	AUDIO     int = 5
	VOICE     int = 6
	VIDEO     int = 7
	VideoNote int = 8
)

var (
	// Contexts
	tdCtx = context.TODO()
	bgCtx = context.Background()

	// define collections
	adminSettingsColl      *mongo.Collection
	blacklistsColl         *mongo.Collection
	pinColl                *mongo.Collection
	userColl               *mongo.Collection
	reportChatColl         *mongo.Collection
	reportUserColl         *mongo.Collection
	devsColl               *mongo.Collection
	chatColl               *mongo.Collection
	channelColl            *mongo.Collection
	antifloodSettingsColl  *mongo.Collection
	connectionColl         *mongo.Collection
	connectionSettingsColl *mongo.Collection
	disableColl            *mongo.Collection
	rulesColl              *mongo.Collection
	warnSettingsColl       *mongo.Collection
	warnUsersColl          *mongo.Collection
	greetingsColl          *mongo.Collection
	lockColl               *mongo.Collection
	filterColl             *mongo.Collection
	notesColl              *mongo.Collection
	notesSettingsColl      *mongo.Collection
)

// dbInstance initializes the MongoDB client and collections
func init() {
	// Create a new MongoDB client
	mongoClient, err := mongo.Connect(tdCtx, options.Client().ApplyURI(config.DatabaseURI))
	if err != nil {
		log.Fatalf("[Database][Connect]: %v", err)
	}

	err = mongoClient.Ping(tdCtx, nil)
	if err != nil {
		log.Fatalf("[Database][Ping]: %v", err)
	}

	// Get the database reference
	db := mongoClient.Database(config.MainDbName)

	// Initialize collections
	log.Info("Opening Database Collections...")
	adminSettingsColl = db.Collection("admin")
	blacklistsColl = db.Collection("blacklists")
	pinColl = db.Collection("pins")
	userColl = db.Collection("users")
	reportChatColl = db.Collection("report_chat_settings")
	reportUserColl = db.Collection("report_user_settings")
	devsColl = db.Collection("devs")
	chatColl = db.Collection("chats")
	channelColl = db.Collection("channels")
	antifloodSettingsColl = db.Collection("antiflood_settings")
	connectionColl = db.Collection("connection")
	connectionSettingsColl = db.Collection("connection_settings")
	disableColl = db.Collection("disable")
	rulesColl = db.Collection("rules")
	warnSettingsColl = db.Collection("warns_settings")
	warnUsersColl = db.Collection("warns_users")
	greetingsColl = db.Collection("greetings")
	lockColl = db.Collection("locks")
	filterColl = db.Collection("filters")
	notesColl = db.Collection("notes")
	notesSettingsColl = db.Collection("notes_settings")

	log.Info("Done opening all database collections!")
}

// updateOne func to update one document
func updateOne(collecion *mongo.Collection, filter bson.M, data interface{}) (err error) {
	_, err = collecion.UpdateOne(tdCtx, filter, bson.M{"$set": data}, options.Update().SetUpsert(true))
	if err != nil {
		log.Errorf("[Database][updateOne]: %v", err)
	}
	return
}

// findOne func to find one document
func findOne(collecion *mongo.Collection, filter bson.M) (res *mongo.SingleResult) {
	res = collecion.FindOne(tdCtx, filter)
	return
}

// countDocs func to count documents
func countDocs(collecion *mongo.Collection, filter bson.M) (count int64, err error) {
	count, err = collecion.CountDocuments(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][countDocs]: %v", err)
	}
	return
}

// findAll func to find all documents
func findAll(collecion *mongo.Collection, filter bson.M) (cur *mongo.Cursor) {
	cur, err := collecion.Find(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][findAll]: %v", err)
	}
	return
}

// deleteOne func to delete one document
func deleteOne(collecion *mongo.Collection, filter bson.M) (err error) {
	_, err = collecion.DeleteOne(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][deleteOne]: %v", err)
	}
	return
}

// deleteMany func to delete many documents
func deleteMany(collecion *mongo.Collection, filter bson.M) (err error) {
	_, err = collecion.DeleteMany(tdCtx, filter)
	if err != nil {
		log.Errorf("[Database][deleteMany]: %v", err)
	}
	return
}
