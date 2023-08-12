package telegramBot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *TelegramBot) startTasks(updates tgbotapi.UpdatesChannel) {
	go func() {
		for update := range updates {
			if update.Message != nil {
				if update.Message.Chat.IsPrivate() {
					// private
					if update.Message.IsCommand() && b.privateCommandReceivedCallback != nil {
						b.privateCommandReceivedCallback(update.Message)
						continue
					}
					if update.Message.ReplyToMessage != nil && b.privateReplyReceivedCallback != nil {
						b.privateReplyReceivedCallback(update.Message)
						continue
					}
					if b.privateMessageReceivedCallback != nil {
						b.privateMessageReceivedCallback(update.Message)
					}
				} else {
					// public
					if update.Message.IsCommand() && b.publicCommandReceivedCallback != nil {
						b.publicCommandReceivedCallback(update.Message)
						continue
					}
					if b.publicMessageReceivedCallback != nil {
						b.publicMessageReceivedCallback(update.Message)
					}
				}
			}
			if update.CallbackQuery != nil {
				if b.callbackReceivedCallback != nil {
					b.callbackReceivedCallback(update.CallbackQuery)
				}
			}
		}
	}()
}
