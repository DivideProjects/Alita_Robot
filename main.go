package main

import (
	"embed"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/config"
	"github.com/divideprojects/Alita_Robot/alita/i18n"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita"
)

//go:embed locales
var Locales embed.FS

func main() {
	// logs if bot is running in debug mode or not
	if config.Debug {
		log.Info("Running in DEBUG Mode...")
	} else {
		log.Info("Running in RELEASE Mode...")
	}

	// Load Locales
	i18n.LoadLocaleFiles(&Locales, "locales")

	// create a new bot
	b, err := gotgbot.NewBot(
		config.BotToken,
		&gotgbot.BotOpts{
			RequestOpts: &gotgbot.RequestOpts{
				APIURL: config.ApiServer,
			},
		},
	)
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	// some initial checks before running bot
	alita.InitialChecks(b)

	// Create updater and dispatcher.
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If a handler returns an error, log it and continue going.
		Error: func(_ *gotgbot.Bot, _ *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil) // create updater with dispatcher

	if _, err = b.DeleteWebhook(nil); err != nil {
		log.Fatalf("[Polling] Failed to remove webhook: %v", err)
	}
	log.Info("[Polling] Removed Webhook!")

	// start the bot in polling mode
	err = updater.StartPolling(b,
		&ext.PollingOpts{
			DropPendingUpdates: config.DropPendingUpdates,
			GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
				AllowedUpdates: config.AllowedUpdates,
			},
		},
	)
	if err != nil {
		log.Fatalf("[Polling] Failed to start polling: %v", err)
	}

	log.Info("[Polling] Started Polling...!")

	// Log the message that bot started
	log.Infof("[Bot] %s has been started...", b.Username)

	// Set Commands of Bot
	log.Info("Setting Custom Commands for PM...!")
	_, err = b.SetMyCommands(
		[]gotgbot.BotCommand{
			{Command: "start", Description: "Starts the Bot"},
			{Command: "help", Description: "Check Help section of bot"},
		},
		&gotgbot.SetMyCommandsOpts{
			Scope:        gotgbot.BotCommandScopeAllPrivateChats{},
			LanguageCode: "en",
		},
	)
	if err != nil {
		log.Fatalf("[Bot] Failed to set commands: %v", err)
	}

	// Loading Modules
	alita.LoadModules(dispatcher)

	// list modules from modules dir
	log.Infof(
		fmt.Sprintf(
			"[Modules] Loaded modules: %s", alita.ListModules(),
		),
	)

	// send a message to log group
	_, err = b.SendMessage(config.MessageDump,
		fmt.Sprintf("<b>Started Bot!</b>\n<b>Mode:</b> %s\n<b>Loaded Modules:</b>\n%s", config.WorkingMode, alita.ListModules()),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
		},
	)
	if err != nil {
		log.Fatalf("[Bot] Failed to send message to log group: %v", err)
	}

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}
