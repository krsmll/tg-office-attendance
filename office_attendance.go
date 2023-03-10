package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type TGBot struct {
	api              *tgbotapi.BotAPI
	updatesChannelID int64
}

func initTGBot() *TGBot {
	token := os.Getenv("TG_BOT_TOKEN")
	channelID, err := strconv.ParseInt(os.Getenv("TG_CHANNEL_ID"), 10, 64)
	if err != nil {
		log.Panic(err)
	}

	tgAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	return &TGBot{
		api:              tgAPI,
		updatesChannelID: channelID,
	}
}

func (t *TGBot) send(c tgbotapi.Chattable) {
	_, err := t.api.Send(c)
	if err != nil {
		log.Println(err)
		return
	}
}

func (t *TGBot) sendDailyMessage() {
	now := time.Now()
	if weekday := now.Weekday(); weekday == time.Friday || weekday == time.Saturday {
		return
	}

	tmrw := now.Add(24 * time.Hour)
	msg := tgbotapi.NewMessage(t.updatesChannelID,
		fmt.Sprintf("Tomorrow is %s, %s.\n\nReact to this message with anything if you will be attending the office tomorrow!",
			tmrw.Format("02.01.2006"), tmrw.Weekday().String()))

	t.send(msg)
}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	loadEnv()
	bot := initTGBot()

	log.Printf("Authorized on account %s", bot.api.Self.UserName)

	c := cron.New()

	// Explanation on cron spec can be found here: https://en.wikipedia.org/wiki/Cron#Overview
	_, err := c.AddFunc("0 21 * * *", bot.sendDailyMessage)
	if err != nil {
		log.Panic(err)
	}

	c.Start()

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt)
	<-sigC
}
