package telegramBot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	TelegramMessageCallbackFunc       func(message *tgbotapi.Message)
	TelegramCallbackQueryCallbackFunc func(callback *tgbotapi.CallbackQuery)
)

type TelegramBot struct {
	tgBot *tgbotapi.BotAPI

	privateCommandReceivedCallback TelegramMessageCallbackFunc
	publicCommandReceivedCallback  TelegramMessageCallbackFunc
	privateMessageReceivedCallback TelegramMessageCallbackFunc
	publicMessageReceivedCallback  TelegramMessageCallbackFunc
	privateReplyReceivedCallback   TelegramMessageCallbackFunc
	callbackReceivedCallback       TelegramCallbackQueryCallbackFunc
}

func NewTelegramBot(botToken string, active bool) (*TelegramBot, error) {
	tgBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := tgBot.GetUpdatesChan(u)

	b := &TelegramBot{
		tgBot:                          tgBot,
		privateCommandReceivedCallback: nil,
		publicCommandReceivedCallback:  nil,
		privateMessageReceivedCallback: nil,
		publicMessageReceivedCallback:  nil,
		privateReplyReceivedCallback:   nil,
		callbackReceivedCallback:       nil,
	}
	if active {
		b.startTasks(updates)
	}

	return b, nil
}

func (b *TelegramBot) SetPrivateCommandReceivedCallback(f TelegramMessageCallbackFunc) {
	b.privateCommandReceivedCallback = f
}

func (b *TelegramBot) SetPublicCommandReceivedCallback(f TelegramMessageCallbackFunc) {
	b.publicCommandReceivedCallback = f
}

func (b *TelegramBot) SetPrivateMessageReceivedCallback(f TelegramMessageCallbackFunc) {
	b.privateMessageReceivedCallback = f
}

func (b *TelegramBot) SetPublicMessageReceivedCallback(f TelegramMessageCallbackFunc) {
	b.publicMessageReceivedCallback = f
}

func (b *TelegramBot) SetPrivateReplyReceivedCallback(f TelegramMessageCallbackFunc) {
	b.privateReplyReceivedCallback = f
}

func (b *TelegramBot) SetCallbackReceivedCallback(f TelegramCallbackQueryCallbackFunc) {
	b.callbackReceivedCallback = f
}

func (b *TelegramBot) SendMessage(chatID int64, text string) (tgbotapi.Message, error) {
	return b.SendFormattedMessage(chatID, text, "")
}

func (b *TelegramBot) SendFormattedMessage(chatID int64, text string, format string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = format
	msg.DisableWebPagePreview = true

	return b.tgBot.Send(msg)
}

func (b *TelegramBot) SendRaw(msg tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.tgBot.Send(msg)
}
