package db

import (
	"errors"
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/dustin/go-humanize"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTeamMemInfo(userID int64) (devrc *Team) {
	defaultTeamMember := &Team{UserId: userID, Dev: false, Sudo: false}
	err := findOne(devsColl, bson.M{"_id": userID}).Decode(&devrc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		devrc = defaultTeamMember
	} else if err != nil {
		devrc = defaultTeamMember
		log.Errorf("[Database] GetTeamMemInfo: %v - %d", err, userID)
	}
	log.Infof("[Database] GetTeamMemInfo: %d", userID)
	return
}

func GetTeamMembers() map[int64]string {
	var teamArray []*Team
	array := make(map[int64]string)
	cursor := findAll(devsColl, bson.M{})
	defer cursor.Close(bgCtx)
	err := cursor.All(bgCtx, &teamArray)
	if err != nil {
		log.Error(err)
		return nil
	}

	for _, result := range teamArray {
		var uPerm string
		if result.Dev {
			uPerm = "dev"
		} else if result.Sudo {
			uPerm = "sudo"
		}
		array[result.UserId] = uPerm
	}

	return array
}

func AddDev(userID int64) {
	var sudo bool
	memInfo := GetTeamMemInfo(userID)
	if memInfo.Dev {
		sudo = false
	}
	teamUpdate := &Team{UserId: userID, Dev: true, Sudo: sudo}
	err := updateOne(devsColl, bson.M{"_id": userID}, teamUpdate)
	if err != nil {
		log.Errorf("[Database] AddDev: %v - %d", err, userID)
		return
	}
	log.Infof("[Database] AddDev: %d", userID)
}

func RemDev(userID int64) {
	err := deleteOne(devsColl, bson.M{"_id": userID})
	if err != nil {
		log.Errorf("[Database] RemDev: %v - %d", err, userID)
	}
}

func LoadAllStats() string {
	totalUsers := LoadUsersStats()
	activeChats, inactiveChats := LoadChatStats()
	AcCount, ClCount := LoadPinStats()
	uRCount, gRCount := LoadReportStats()
	antiCount := LoadAntifloodStats()
	setRules, pvtRules := LoadRulesStats()
	blacklistTriggers, blacklistChats := LoadBlacklistsStats()
	connectedUsers, connectedChats := LoadConnectionStats()
	disabledCmds, disableEnabledChats := LoadDisableStats()
	filtersNum, filtersChats := LoadFilterStats()
	enabledWelcome, enabledGoodbye, cleanServiceEnabled, cleanWelcomeEnabled, cleanGoodbyeEnabled := LoadGreetingsStats()
	notesNum, notesChats := LoadNotesStats()
	numChannels := LoadChannelStats()

	result := "<u>Alita's Stats:</u>" +
		fmt.Sprintf("\n\nGo Version: %s", runtime.Version()) +
		fmt.Sprintf("\nGoroutines: %s", humanize.Comma(int64(runtime.NumGoroutine()))) +
		fmt.Sprintf("\n<b>Antiflood:</b> enabled in %s chats", humanize.Comma(antiCount)) +
		fmt.Sprintf(
			"\n<b>Users:</b> %s users found in %s active Chats (%s Inactive, %s Total)",
			humanize.Comma(totalUsers),
			humanize.Comma(int64(activeChats)),
			humanize.Comma(int64(inactiveChats)),
			humanize.Comma(int64(activeChats+inactiveChats)),
		) +
		"\n<b>Pins:</b>" +
		fmt.Sprintf("\n    <b>CleanLinked Enabled:</b> %s", humanize.Comma(ClCount)) +
		fmt.Sprintf("\n    <b>AntiChannelPin Enabled:</b> %s", humanize.Comma(AcCount)) +
		fmt.Sprintf(
			"\n<b>Reports:</b> %s users enabled reports in %s Chats",
			humanize.Comma(uRCount),
			humanize.Comma(gRCount),
		) +
		"\n<b>Rules:</b>" +
		fmt.Sprintf("\n    <b>Set:</b> %s", humanize.Comma(setRules)) +
		fmt.Sprintf("\n    <b>Private:</b> %s", humanize.Comma(pvtRules)) +
		fmt.Sprintf(
			"\n<b>Blacklists:</b> %s triggers in %s chats",
			humanize.Comma(blacklistTriggers),
			humanize.Comma(blacklistChats),
		) +
		"\n<b>Connections:</b>" +
		fmt.Sprintf("\n    %s users connected to chats", humanize.Comma(connectedUsers)) +
		fmt.Sprintf("\n    %s chats allow user connections", humanize.Comma(connectedChats)) +
		fmt.Sprintf(
			"\n<b>Disabling:</b> %s commands disabled in %s chats",
			humanize.Comma(disabledCmds),
			humanize.Comma(disableEnabledChats),
		) +
		fmt.Sprintf(
			"\n<b>Filters:</b> %s filters saved in %s chats",
			humanize.Comma(filtersNum),
			humanize.Comma(filtersChats),
		) +
		"\n<b>Greetings:</b>" +
		fmt.Sprintf("\n    <b>Welcome Enabled:</b> %s", humanize.Comma(enabledWelcome)) +
		fmt.Sprintf("\n    <b>Goodbye Enabled:</b> %s", humanize.Comma(enabledGoodbye)) +
		fmt.Sprintf("\n    <b>CleanService:</b> %s", humanize.Comma(cleanServiceEnabled)) +
		fmt.Sprintf("\n    <b>CleanWelcome:</b> %s", humanize.Comma(cleanWelcomeEnabled)) +
		fmt.Sprintf("\n    <b>CleanGoodbye:</b> %s", humanize.Comma(cleanGoodbyeEnabled)) +
		fmt.Sprintf(
			"\n<b>Notes:</b> %s notes saved in %s chats",
			humanize.Comma(notesNum),
			humanize.Comma(notesChats),
		) +
		fmt.Sprintf("\n<b>Channels Stored</b>: %s", humanize.Comma(numChannels))

	return result
}

type Team struct {
	UserId int64 `bson:"_id,omitempty" json:"_id,omitempty"`
	Dev    bool  `bson:"dev" json:"dev" default:"false"`
	Sudo   bool  `bson:"sudo" json:"sudo" default:"false"`
}

func AddSudo(userID int64) {
	var dev bool
	memInfo := GetTeamMemInfo(userID)
	if memInfo.Dev {
		dev = false
	}
	teamUpdate := &Team{UserId: userID, Dev: dev, Sudo: true}
	err := updateOne(devsColl, bson.M{"_id": userID}, teamUpdate)
	if err != nil {
		log.Errorf("[Database] AddSudo: %v - %d", err, userID)
		return
	}
	log.Infof("[Database] AddSudo: %d", userID)
}

func RemSudo(userID int64) {
	err := deleteOne(devsColl, bson.M{"_id": userID})
	if err != nil {
		log.Errorf("[Database] RemSudo: %v - %d", err, userID)
		return
	}
}
