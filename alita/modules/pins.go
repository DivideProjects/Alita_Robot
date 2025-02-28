package modules

import (
	"fmt"
	"strings"

	tgmd2html "github.com/PaulSonOfLars/gotg_md2html"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"

	"github.com/divideprojects/Alita_Robot/alita/utils/string_handling"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/chat_status"
)

var pinsModule = moduleStruct{
	moduleName:   "Pins",
	handlerGroup: 10,
}

type pinType struct {
	MsgText  string
	FileID   string
	DataType int
}

/*
	Check new pinned messages

# This function works for AntiChannelPin and CleanLinked

# AntiChannelPin - Unpins message pinned by channel

# CleanLinked - Deletes the message linked by channel

This a watcher for 2 functions described above
*/
func (moduleStruct) checkPinned(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	pinprefs := db.GetPinData(chat.Id)

	if pinprefs.CleanLinked {
		_, err := b.DeleteMessage(chat.Id, msg.MessageId, nil)
		// if err.Error() == "unable to deleteMessage: Bad Request: message to delete not found" {
		// 	log.WithFields(
		// 		log.Fields{
		// 			"chat": chat.Id,
		// 		},
		// 	).Error("error deleting message")
		// 	return ext.EndGroups
		// } else
		if err != nil {
			log.Error(err)
			return err
		}
	} else if pinprefs.AntiChannelPin {
		_, err := b.UnpinChatMessage(chat.Id,
			&gotgbot.UnpinChatMessageOpts{
				MessageId: &msg.MessageId,
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.ContinueGroups
}

/* Unpin the latest pinned message or message to which user replied

This function unpins the latest pinned message or message to which user replied */

func (moduleStruct) unpin(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User

	// Check permissions
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotPin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserPin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}

	var (
		replyText  string
		replyMsgId int64
	)

	if replyMsg := msg.ReplyToMessage; replyMsg != nil {
		replyMsgId = replyMsg.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	if rMsg := msg.ReplyToMessage; rMsg != nil {
		replyText = "Unpinned this message."
		_, err := b.UnpinChatMessage(chat.Id, &gotgbot.UnpinChatMessageOpts{MessageId: &rMsg.MessageId})
		if err != nil {
			replyText = "failed to unpin:" + err.Error()

		}
	} else {
		replyText = "Unpinned the last pinned message."
		_, err := b.UnpinChatMessage(chat.Id, nil)
		if err != nil {
			replyText = "failed to unpin:" + err.Error()
		}
	}

	_, err := msg.Reply(b, replyText,
		&gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{IsDisabled: true},
			ReplyParameters:    &gotgbot.ReplyParameters{MessageId: replyMsgId, AllowSendingWithoutReply: true},
		},
	)

	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// Callback Query Handler for Unpinall command
func (moduleStruct) unpinallCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	query := ctx.Update.CallbackQuery
	chat := ctx.EffectiveChat

	if query.Data == "unpinallbtn(yes)" {
		status, err := b.UnpinAllChatMessages(chat.Id, nil)
		if !status && err != nil {
			log.Errorf("[Pin] UnpinAllChatMessages: %d", chat.Id)
			return err
		}
		_, _, erredit := query.Message.EditText(b, "Unpinned all pinned messages in this chat!", nil)
		if erredit != nil {
			log.Errorf("[Pin] UnpinAllChatMessages: %d", chat.Id)
			return err
		}
	} else if query.Data == "unpinallbtn(no)" {
		_, _, err := query.Message.EditText(b, "Cancelled operation to unpin messages!", nil)
		if err != nil {
			log.Errorf("[Pin] UnpinAllChatMessages: %d", chat.Id)
			return err
		}
	}
	return ext.EndGroups
}

/* Unpin all the pinned messages in the chat

Can only be used by owner to unpin all message in a chat. */

func (moduleStruct) unpinAll(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User

	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotPin(b, ctx, nil, false) {
		return ext.EndGroups
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, "Are you sure you want to unpin all pinned messages?",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "Yes", CallbackData: "unpinallbtn(yes)"},
						{Text: "No", CallbackData: "unpinallbtn(no)"},
					},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}
	return ext.EndGroups
}

/* Bot pins the message followed by command

The users message is pinned by bot and includes buttons as well */

func (m moduleStruct) permaPin(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	user := ctx.EffectiveSender.User
	args := ctx.Args()

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserPin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotPin(b, ctx, nil, false) {
		return ext.EndGroups
	}

	// if command is empty (i.e. Without Arguments) not replied to a message, return and end group
	if len(args) == 1 && msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, "Please reply to a message or give some text to pin.", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	var (
		buttons []tgmd2html.ButtonV2
		pinT    = pinType{}
	)

	pinT.FileID, pinT.MsgText, pinT.DataType, buttons = m.GetPinType(msg)
	if pinT.DataType == -1 {
		_, err := msg.Reply(b, "Permapin not supported for data type!", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	keyb := helpers.BuildKeyboard(helpers.ConvertButtonV2ToDbButton(buttons))

	// enum func works here
	ppmsg, err := PinsEnumFuncMap[pinT.DataType](b, ctx, pinT, &gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyb}, 0)
	if err != nil {
		log.Error(err)
		return err
	}

	msgToPin := ppmsg.MessageId
	pin, err := b.PinChatMessage(chat.Id, msgToPin, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	if pin {
		pinLink := helpers.GetMessageLinkFromMessageId(chat, msgToPin)
		_, err = msg.Reply(b,
			fmt.Sprintf("I have pinned <a href='%s'>this message</a>", pinLink),
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                msgToPin,
					AllowSendingWithoutReply: true,
				},
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
			},
		)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

/* Pins the message replied by user

Normally pins message without tagging users but tag can be
enabled by entering 'notify'/'violent'/'loud' in front of command */

func (moduleStruct) pin(b *gotgbot.Bot, ctx *ext.Context) error {
	user := ctx.EffectiveSender.User
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage
	isSilent := true
	args := ctx.Args

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireUserAdmin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.CanUserPin(b, ctx, nil, user.Id, false) {
		return ext.EndGroups
	}
	if !chat_status.CanBotPin(b, ctx, nil, false) {
		return ext.EndGroups
	}

	if msg.ReplyToMessage == nil {
		_, err := msg.Reply(b, "Reply to a message to pin it", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return ext.EndGroups
	}

	prevMessage := msg.ReplyToMessage
	pinMsg := "I have pinned <a href='%s'>this message</a>."

	if len(args()) > 0 {
		isSilent = !string_handling.FindInStringSlice([]string{"notify", "violent", "loud"}, args()[0])
		if !isSilent {
			pinMsg = "I have pinned <a href='%s'>this message</a> loudly!"
		}
	}

	pin, err := b.PinChatMessage(chat.Id,
		prevMessage.MessageId,
		&gotgbot.PinChatMessageOpts{
			DisableNotification: isSilent,
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	if pin {
		pinLink := helpers.GetMessageLinkFromMessageId(chat, prevMessage.MessageId)
		_, err = prevMessage.Reply(b,
			fmt.Sprintf(
				pinMsg,
				pinLink,
			),
			helpers.Shtml(),
		)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return ext.EndGroups
}

/*
	Enable or Disable AntiChannelPin

connection - true, true

Sets Preference for checkPinned function to check message for unpinning or not
*/
func (moduleStruct) antichannelpin(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()

	if len(args) >= 2 {
		switch strings.ToLower(args[1]) {
		case "on", "yes", "true":
			go db.SetAntiChannelPin(chat.Id, true)
			_, err := msg.Reply(b,
				"<b>Enabled</b> anti channel pins. Automatic pins from a channel will now be replaced with the previous pin.",
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		case "off", "no", "false":
			go db.SetAntiChannelPin(chat.Id, false)
			_, err := msg.Reply(b,
				"<b>Disabled</b> anti channel pins. Automatic pins from a channel will not be removed.",
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		default:
			_, err := msg.Reply(b, "Your input was not recognised as one of: yes/no/on/off", helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else {
		pinprefs := db.GetPinData(chat.Id)
		if pinprefs.AntiChannelPin {
			_, err := msg.Reply(b,
				fmt.Sprintf("Anti-Channel pins are currently <b>enabled</b> in %s. All channel posts that get auto-pinned by telegram will be replaced with the previous pin.", chat.Title),
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			_, err := msg.Reply(b,
				fmt.Sprintf("Anti-Channel pins are currently <b>disabled</b> in %s.", chat.Title),
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return ext.EndGroups
}

/*
	Enable or Disable CleanLinked

connection - true, true

Sets Preference for checkPinned function to check message for cleaning or not
*/
func (moduleStruct) cleanlinked(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage
	// connection status
	connectedChat := helpers.IsUserConnected(b, ctx, true, true)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	args := ctx.Args()

	if len(args) >= 2 {
		switch strings.ToLower(args[1]) {
		case "on", "yes", "true":
			go db.SetCleanLinked(chat.Id, true)
			_, err := msg.Reply(b,
				"<b>Enabled</b> linked channel post deletion in Logs. Messages sent from the linked channel will be deleted.",
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		case "off", "no", "false":
			go db.SetCleanLinked(chat.Id, false)
			_, err := msg.Reply(b,
				"<b>Disabled</b> linked channel post deletion in Logs.",
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		default:
			_, err := msg.Reply(b, "Your input was not recognised as one of: yes/no/on/off", helpers.Shtml())
			if err != nil {
				log.Error(err)
				return err
			}
		}
	} else {
		pinprefs := db.GetPinData(chat.Id)
		if pinprefs.CleanLinked {
			_, err := msg.Reply(b,
				fmt.Sprintf("Linked channel post deletion is currently <b>enabled</b> in %s. Messages sent from the linked channel will be deleted", chat.Title),
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		} else {
			_, err := msg.Reply(b,
				fmt.Sprintf("Linked channel post deletion is currently <b>disabled</b> in %s.", chat.Title),
				helpers.Shtml(),
			)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}

	return ext.EndGroups
}

/*
	Gets the pinned message in chat

connection - false, true

User can get the link to latest pinned message of chat using this
*/
func (moduleStruct) pinned(b *gotgbot.Bot, ctx *ext.Context) error {
	chat := ctx.EffectiveChat
	msg := ctx.EffectiveMessage

	// permission checks
	if !chat_status.RequireGroup(b, ctx, nil, false) {
		return ext.EndGroups
	}
	if !chat_status.RequireBotAdmin(b, ctx, nil, false) {
		return ext.EndGroups
	}

	var (
		pinLink    string
		replyMsgId int64
	)

	if reply := msg.ReplyToMessage; reply != nil {
		replyMsgId = reply.MessageId
	} else {
		replyMsgId = msg.MessageId
	}

	chatInfo, err := b.GetChat(chat.Id, nil)
	if err != nil {
		log.Error(err)
		return err
	}

	pinnedMsg := chatInfo.PinnedMessage

	if pinnedMsg == nil {
		_, err = msg.Reply(b, "No message has been pinned in the current chat!", helpers.Shtml())
		if err != nil {
			log.Error(err)
			return err
		}
		return err
	}

	pinLink = helpers.GetMessageLinkFromMessageId(chat, pinnedMsg.MessageId)

	_, err = msg.Reply(b,
		fmt.Sprintf("<a href='%s'>Here</a> is the pinned message.", pinLink),
		&gotgbot.SendMessageOpts{
			ParseMode: helpers.HTML,
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId:                replyMsgId,
				AllowSendingWithoutReply: true,
			},
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: "Pinned Message", Url: pinLink},
					},
				},
			},
		},
	)
	if err != nil {
		log.Error(err)
		return err
	}

	return ext.EndGroups
}

// PinsEnumFuncMap
// A rather very complicated PinsEnumFuncMap Variable made by me to send filters in an appropriate way
var PinsEnumFuncMap = map[int]func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error){
	db.TEXT: func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error) {
		return b.SendMessage(
			ctx.EffectiveChat.Id,
			pinT.MsgText,
			&gotgbot.SendMessageOpts{
				ParseMode: helpers.HTML,
				LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyMarkup: keyb,
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.STICKER: func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error) {
		return b.SendSticker(
			ctx.EffectiveChat.Id,
			gotgbot.InputFileByID(pinT.FileID),
			&gotgbot.SendStickerOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				ReplyMarkup:     keyb,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.DOCUMENT: func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error) {
		return b.SendDocument(
			ctx.EffectiveChat.Id,
			gotgbot.InputFileByID(pinT.FileID),
			&gotgbot.SendDocumentOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				ParseMode:       helpers.HTML,
				ReplyMarkup:     keyb,
				Caption:         pinT.MsgText,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.PHOTO: func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error) {
		return b.SendPhoto(
			ctx.EffectiveChat.Id,
			gotgbot.InputFileByID(pinT.FileID),
			&gotgbot.SendPhotoOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				ParseMode:       helpers.HTML,
				ReplyMarkup:     keyb,
				Caption:         pinT.MsgText,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.AUDIO: func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error) {
		return b.SendAudio(
			ctx.EffectiveChat.Id,
			gotgbot.InputFileByID(pinT.FileID),
			&gotgbot.SendAudioOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				ParseMode:       helpers.HTML,
				ReplyMarkup:     keyb,
				Caption:         pinT.MsgText,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.VOICE: func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error) {
		return b.SendVoice(
			ctx.EffectiveChat.Id,
			gotgbot.InputFileByID(pinT.FileID),
			&gotgbot.SendVoiceOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				ParseMode:       helpers.HTML,
				ReplyMarkup:     keyb,
				Caption:         pinT.MsgText,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
	db.VIDEO: func(b *gotgbot.Bot, ctx *ext.Context, pinT pinType, keyb *gotgbot.InlineKeyboardMarkup, replyMsgId int64) (*gotgbot.Message, error) {
		return b.SendVideo(
			ctx.EffectiveChat.Id,
			gotgbot.InputFileByID(pinT.FileID),
			&gotgbot.SendVideoOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId:                replyMsgId,
					AllowSendingWithoutReply: true,
				},
				ParseMode:       helpers.HTML,
				ReplyMarkup:     keyb,
				Caption:         pinT.MsgText,
				MessageThreadId: ctx.EffectiveMessage.MessageThreadId,
			},
		)
	},
}

func (moduleStruct) GetPinType(msg *gotgbot.Message) (fileid, text string, dataType int, buttons []tgmd2html.ButtonV2) {
	dataType = -1 // not defined datatype; invalid filter
	var (
		rawText string
		args    = strings.Split(msg.Text, " ")[1:]
	)

	if reply := msg.ReplyToMessage; reply != nil {
		if reply.Text == "" {
			rawText = reply.OriginalCaptionMDV2()
		} else {
			rawText = reply.OriginalMDV2()
		}
	} else {
		if msg.Text == "" {
			rawText = strings.SplitN(msg.OriginalCaptionMDV2(), " ", 2)[1]
		} else {
			rawText = strings.SplitN(msg.OriginalMDV2(), " ", 2)[1]
		}
	}

	// get text and buttons
	text, buttons = tgmd2html.MD2HTMLButtonsV2(rawText)

	if len(args) >= 1 && msg.ReplyToMessage == nil {
		dataType = db.TEXT
	} else if msg.ReplyToMessage != nil && len(args) >= 0 {
		if len(args) >= 0 && msg.ReplyToMessage.Text != "" {
			dataType = db.TEXT
		} else if msg.ReplyToMessage.Sticker != nil {
			fileid = msg.ReplyToMessage.Sticker.FileId
			dataType = db.STICKER
		} else if msg.ReplyToMessage.Document != nil {
			fileid = msg.ReplyToMessage.Document.FileId
			dataType = db.DOCUMENT
		} else if len(msg.ReplyToMessage.Photo) > 0 {
			fileid = msg.ReplyToMessage.Photo[len(msg.ReplyToMessage.Photo)-1].FileId // using -1 index to get best photo quality
			dataType = db.PHOTO
		} else if msg.ReplyToMessage.Audio != nil {
			fileid = msg.ReplyToMessage.Audio.FileId
			dataType = db.AUDIO
		} else if msg.ReplyToMessage.Voice != nil {
			fileid = msg.ReplyToMessage.Voice.FileId
			dataType = db.VOICE
		} else if msg.ReplyToMessage.Video != nil {
			fileid = msg.ReplyToMessage.Video.FileId
			dataType = db.VIDEO
		}
	}

	return
}

func LoadPin(dispatcher *ext.Dispatcher) {
	HelpModule.AbleMap.Store(pinsModule.moduleName, true)

	dispatcher.AddHandler(handlers.NewCommand("unpin", pinsModule.unpin))
	dispatcher.AddHandler(handlers.NewCommand("unpinall", pinsModule.unpinAll))
	dispatcher.AddHandler(handlers.NewCommand("pin", pinsModule.pin))
	dispatcher.AddHandler(handlers.NewCommand("pinned", pinsModule.pinned))
	dispatcher.AddHandler(handlers.NewCommand("antichannelpin", pinsModule.antichannelpin))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("unpinallbtn"), pinsModule.unpinallCallback))
	dispatcher.AddHandlerToGroup(
		handlers.NewMessage(
			func(msg *gotgbot.Message) bool {
				return msg.GetSender().IsLinkedChannel()
			},
			pinsModule.checkPinned,
		),
		pinsModule.handlerGroup,
	)
	dispatcher.AddHandler(handlers.NewCommand("permapin", pinsModule.permaPin))
	dispatcher.AddHandler(handlers.NewCommand("cleanlinked", pinsModule.cleanlinked))
}
