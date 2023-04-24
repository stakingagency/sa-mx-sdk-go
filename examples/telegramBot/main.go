package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/stakingagency/sa-mx-sdk-go/telegramBot"
)

const (
	botToken = "6015747887:AAE3HXUaOQPhWBAsJq8JsT9oQnRMPqWGCV0" // @SdkTestBot
)

var bot *telegramBot.TelegramBot

func main() {
	var err error
	bot, err = telegramBot.NewTelegramBot(botToken, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	bot.SetPrivateCommandReceivedCallback(privateCommandReceived)
	for {
	}
}

func privateCommandReceived(message *tgbotapi.Message) {
	command := message.Command()
	from := fmt.Sprintf("@%s (%s %s)", message.From.UserName, message.From.FirstName, message.From.LastName)
	fmt.Printf("%s command received from %s\n", command, from)

	text := fmt.Sprintf("you sent me the `%s` command", command)
	_, err := bot.SendFormattedMessage(int64(message.From.ID), text, tgbotapi.ModeMarkdown)
	if err != nil {
		fmt.Println(err)
	}
}
