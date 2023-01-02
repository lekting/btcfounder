package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	BOT_TOKEN  string
	BOT_CHAT_ID string
)

func InitBot() {
	BOT_TOKEN   = os.Getenv("BOT_TOKEN")
	BOT_CHAT_ID = os.Getenv("USER_ID")
}

func SendBotMessage(message string) {
	data := url.Values{
		"chat_id": {BOT_CHAT_ID},
		"text":    {message},
	}

	_, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", BOT_TOKEN), data)

	if err != nil {
		log.Fatal(err)
	}
}
