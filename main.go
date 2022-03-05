package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kbinani/screenshot"
	"github.com/tkanos/gonfig"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

type Configuration struct {
	Token  string
	UserId int64
}

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/send"),
		tgbotapi.NewKeyboardButton("/clip"),
	),
)

func main() {
	config := initBot()
	fmt.Println(config)
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}
		chatId := update.Message.Chat.ID
		if (config.UserId != 0) && !(chatId == config.UserId) {
			continue
		}

		msg := tgbotapi.NewMessage(chatId, "")

		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand /send and /clip."
		case "send":
			sendPhoto(chatId, bot)
			continue
		case "clip":
			msg.Text, msg.ReplyMarkup = clipString()
			msg.Text += "\n\nFrom: " + getUsername()
			msg.ParseMode = "markdown"
		case "start":
			msg.Text = "Bot Started 😁"
			msg.ReplyMarkup = numericKeyboard
		case "close":
			msg.Text = "Shortcut Closed"
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		case "exit":
			os.Exit(0)
		default:
			msg.Text = "I don't know that command" + getUsername()
		}
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func initBot() (config Configuration) {
	err := gonfig.GetConf("config.json", &config)
	if err != nil {
		newConfig := Configuration{
			Token:  "YOUR_TOKEN_HERE",
			UserId: 0,
		}
		file, _ := json.MarshalIndent(newConfig, "", " ")
		_ = ioutil.WriteFile("config.json", file, 0644)
		os.Exit(1)
	}
	return
}

func sendPhoto(chatId int64, bot *tgbotapi.BotAPI) {
	img, err := screenshot.CaptureDisplay(0)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
	photoBytes := buf.Bytes()

	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: photoBytes,
	}
	msg := tgbotapi.NewPhoto(chatId, photoFileBytes)
	msg.Caption = "From: " + getUsername()

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func clipString() (string, tgbotapi.InlineKeyboardMarkup) {
	text, err := clipboard.ReadAll()
	if err != nil {
		text = "Something Went Wrong!!!"
	}
	clipButton := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Search Google", "https://www.google.com/search?q="+text),
		),
	)
	return fmt.Sprintf("`%s`", text), clipButton
}

func getUsername() string {
	userData, err := user.Current()
	if err != nil {
		userData.Name = "Anonymous"
	}
	return userData.Name
}
