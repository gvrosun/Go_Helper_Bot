//go:generate goversioninfo -icon=exam.ico -manifest=goversioninfo.exe.manifest
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kbinani/screenshot"
	"github.com/nfnt/resize"
	"github.com/tkanos/gonfig"
	"gopkg.in/toast.v1"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
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
	if !connected() {
		toastNotification("No Internet Connection", "Please Check your connectivity and try again!")
		os.Exit(1)
	}
	config := initBot()
	if config.Token == "YOUR_TOKEN_HERE" {
		toastNotification("Attention!!!", "Please provide your Telegram Bot API Token in config.json file")
	}
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		toastNotification("Error!!!", "Please verify your Token and try again.")
		os.Exit(1)
	}
	toastNotification("Bot Initialed...", "Now you are good to go, Enter /start in telegram bot.")

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
		if (config.UserId != 0) && !(update.Message.From.ID == config.UserId) {
			continue
		}

		msg := tgbotapi.NewMessage(chatId, "")

		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand these commands:\n" +
				"/start - Start Bot and Enable Shortcut\n" +
				"/send - Send captured Screenshot\n" +
				"/clip - Send Clipboard data\n" +
				"/close - Close the Shortcut buttons\n" +
				"/exit - Stop Bot Application completely\n" +
				"/help - Know more about each commands"
		case "send":
			sendPhoto(chatId, bot)
			continue
		case "clip":
			msg.Text, msg.ReplyMarkup = clipString()
			msg.Text += "\n\nFrom: " + getUsername()
			msg.ParseMode = "markdown"
		case "start":
			msg.Text = "Bot Working...\n" +
				"List of Commands:\n" +
				"/start - Start Bot and Enable Shortcut\n" +
				"/send - Send captured Screenshot\n" +
				"/clip - Send Clipboard data\n" +
				"/close - Close the Shortcut buttons\n" +
				"/exit - Stop Bot Application completely\n" +
				"/help - Know more about each commands"
			msg.ReplyMarkup = numericKeyboard
		case "close":
			msg.Text = "Shortcut Closed"
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		case "exit":
			toastNotification("Bot Terminated...", "Bot exited and cleared from memory...")
			os.Exit(0)
		default:
			msg.Text = "I don't know that command"
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
		toastNotification("Attention Needed!!!", "Config.json file is generated, Please review the file to fill the details required")
		os.Exit(1)
	}
	return
}

func sendPhoto(chatId int64, bot *tgbotapi.BotAPI) {
	img, err := screenshot.CaptureDisplay(0)
	resizeImg := resize.Resize(1280, 720, img, resize.Lanczos3)
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, resizeImg, nil)
	photoBytes := buf.Bytes()

	photoFileBytes := tgbotapi.FileBytes{
		Name:  "ScreenShot",
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
	text = strings.Replace(text, "\"", "'", -1)
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

func toastNotification(title string, message string) {
	notification := toast.Notification{
		AppID:   "Go Helper",
		Title:   title,
		Message: message,
		Icon:    "R:\\Development\\Go\\Helper\\exam.png",
		Audio:   toast.Reminder,
		Actions: []toast.Action{
			{"protocol", "Exit", ""},
		},
	}
	err := notification.Push()
	if err != nil {
		log.Fatalln(err)
	}
}

func connected() (ok bool) {
	_, err := http.Get("http://clients3.google.com/generate_204")
	if err != nil {
		return false
	}
	return true
}
