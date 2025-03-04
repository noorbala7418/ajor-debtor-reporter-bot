package main

import (
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/noorbala7418/ajor-debtor-reporter-bot/pkg/xray"
	"github.com/sirupsen/logrus"
)

var tgdebug bool = false

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
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	bot.Debug = tgdebug

	logrus.Info("Registered On BOT: ", bot.Self.UserName)
	logrus.Info("Admin IDs: ", os.Getenv("TELEGRAM_BOT_ADMIN_ID"))
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
			if isAdmin(update) {
				msg.Text = `Commands:
				- /all (number) -> for get bulk users (under 60) - Default is 50
				- /debtor
				- /disabled
				- /configs prefix
				- /status YOUR_UID
				- /s YOUR_UID
				- /almost (number) -> Get configs that will be over soon (Default is 1GB).
				`
			} else {
				msg.Text = `
				You can use /status command to know about your config. Example: 
				- /status XXXXXX
				`
			}
		case "start":
			msg.Text = `
			Welcome to Ajor Debtor Reporter Bot 🧱
Use /help command to know about this bot.
			`
		case "disabled":
			if isAdmin(update) {
				msg.ParseMode = "markdown"
				msg.Text = xray.GetDisabledClients()
			} else {
				msg.Text = "Access Denied."
			}
		case "debtor":
			if isAdmin(update) {
				msg.ParseMode = "markdown"
				msg.Text = xray.GetDepletedClients()
			} else {
				msg.Text = "Access Denied."
			}
		case "all":
			if isAdmin(update) {
				msg.ParseMode = "markdown"
				blockPart, _ := strconv.Atoi(strings.Split(update.Message.CommandArguments(), " ")[0])
				result := xray.GetAllClients(blockPart)
				if result == nil {
					msg.Text = "Empty."
				}
				for item := range result {
					msg.Text = result[item]
					if _, err := bot.Send(msg); err != nil {
						logrus.Error("Error in send message to telegram", err)
					}
				}
				continue
			} else {
				msg.Text = "Access Denied."
			}
		case "status":
			msg.ParseMode = "markdown"
			msg.Text = xray.GetSingleConfigStatus(strings.Split(update.Message.CommandArguments(), " ")[0])
		case "s":
			msg.ParseMode = "markdown"
			msg.Text = xray.GetSingleConfigStatus(strings.Split(update.Message.CommandArguments(), " ")[0])
		case "configs":
			if isAdmin(update) {
				msg.ParseMode = "markdown"
				msg.Text = xray.GetConfigsWithPrefix(strings.Split(update.Message.CommandArguments(), " ")[0])
			} else {
				msg.Text = "Access Denied."
			}
		case "almost":
			if isAdmin(update) {
				msg.ParseMode = "markdown"
				limit, _ := strconv.Atoi(strings.Split(update.Message.CommandArguments(), " ")[0])
				msg.Text = xray.GetConfigsAlmostOver(limit)
			} else {
				msg.Text = "Access Denied."
			}
		default:
			msg.ParseMode = "markdown"
			msg.Text = "Command not found."
		}

		if _, err := bot.Send(msg); err != nil {
			logrus.Error("Error in send message to telegram", err)
		}
	}
}

// checkEnvs Checks environment variables and if one variable does not exist, Then it will Kill application.
func checkEnvs() {
	if os.Getenv("TELEGRAM_BOT_ADMIN_ID") == "" {
		logrus.Error("env variable $TELEGRAM_BOT_ADMIN_ID is not defined")
		os.Exit(1)
	}

	if os.Getenv("TELEGRAM_BOT_DEBUG_MODE") == "" {
		logrus.Warning("env variable $TELEGRAM_BOT_DEBUG_MODE is not defined. Default is False")
	} else {
		tgdebug, _ = strconv.ParseBool(os.Getenv("TELEGRAM_BOT_DEBUG_MODE"))
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

// isAdmin returns True when user ID equals to one of admin IDs.
func isAdmin(update tgbotapi.Update) bool {
	adminIDs := strings.Split(os.Getenv("TELEGRAM_BOT_ADMIN_ID"), ",")
	for i := 0; i < len(adminIDs); i++ {
		id, _ := strconv.ParseInt(adminIDs[i], 10, 64)
		if update.Message.Chat.ID == id {
			logrus.Debug("Got admin request from ", adminIDs[i])
			return true
		}
	}
	return false
}
