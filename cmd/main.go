package main

import (
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/noorbala7418/ajor-debtor-reporter-bot/pkg/xray"
	"github.com/sirupsen/logrus"
)

var tgadminid int64
var tgdebug bool

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only logrus the warning severity or above.
	switch os.Getenv("APP_LOG_MODE") {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	checkEnvs()
	DefineEnvs()
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	bot.Debug = tgdebug

	logrus.Info("Registered On BOT: ", bot.Self.UserName)
	logrus.Info("Admin ID: ", tgadminid)
	logrus.Info("DEBUG MODE: ", tgdebug)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			if checkAdmin(update, bot, tgadminid) {
				msg.Text = `Commands:
				- /debtor
				- /all
				- /status YOUR_UID
				`
			} else {
				msg.Text = `
				You can use /status command to know about your config. Example: 
				- /status XXXXXX
				`
			}
		case "start":
			msg.Text = "Use /help command to know about this bot."
		case "debtor":
			if checkAdmin(update, bot, tgadminid) {
				msg.ParseMode = "markdown"
				msg.Text = xray.GetDepletedClients()
			} else {
				msg.Text = "Access Denied."
			}
		case "all":
			if checkAdmin(update, bot, tgadminid) {
				msg.ParseMode = "markdown"
				msg.Text = xray.GetAllClients()
			} else {
				msg.Text = "Access Denied."
			}
		case "status":
			msg.ParseMode = "markdown"
			// xray.GetSingleConfigStatus(strings.Split(update.Message.CommandArguments(), " ")[0])
			// msg.Text = "Not implemented."
			msg.Text = xray.GetSingleConfigStatus(strings.Split(update.Message.CommandArguments(), " ")[0])
		default:
			msg.ParseMode = "markdown"
			msg.Text = "Command not found."
		}

		if _, err := bot.Send(msg); err != nil {
			logrus.Error("Error in send message to telegram", err)
		}
	}
}

func checkEnvs() {
	if os.Getenv("TELEGRAM_BOT_ADMIN_ID") == "" {
		logrus.Error("env variable $TELEGRAM_BOT_ADMIN_ID is not defined")
		os.Exit(1)
	}

	if os.Getenv("TELEGRAM_BOT_DEBUG_MODE") == "" {
		logrus.Error("env variable $TELEGRAM_BOT_DEBUG_MODE is not defined")
		os.Exit(1)
	}

	if os.Getenv("TELEGRAM_BOT_TOKEN") == "" {
		logrus.Error("env variable $TELEGRAM_BOT_TOKEN is not defined")
		os.Exit(1)
	}

	if os.Getenv("XPANEL_URL") == "" {
		logrus.Error("env variable $XPANEL_URL is not defined")
		os.Exit(1)
	}

	if os.Getenv("XPANEL_USERNAME") == "" {
		logrus.Error("env variable $XPANEL_USERNAME is not defined")
		os.Exit(1)
	}

	if os.Getenv("XPANEL_PASSWORD") == "" {
		logrus.Error("env variable $XPANEL_PASSWORD is not defined")
		os.Exit(1)
	}
}

func DefineEnvs() {
	tgadminid, _ = strconv.ParseInt(os.Getenv("TELEGRAM_BOT_ADMIN_ID"), 10, 64)
	tgdebug, _ = strconv.ParseBool(os.Getenv("TELEGRAM_BOT_DEBUG_MODE"))
}

func checkAdmin(update tgbotapi.Update, bot *tgbotapi.BotAPI, tgAdminID int64) bool {

	return update.Message.Chat.ID == tgAdminID
}
