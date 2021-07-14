package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"time"
)

type Response struct {
	CreatedAt time.Time `json:"created_at"`
	Url       string    `json:"url"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	Body      string    `json:"body"`
}

const UpdateCycle = 5
const (
	BotToken = "BOT_TOKEN"
	Token    = "GIT_TOKEN"
	User     = "GIT USER"
	Repo     = "GIT REPO"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic("Cannot initialize the bot: " + err.Error())
	}
	log.Println("Github Issues bot for Telegram")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		} else if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to Github Issues bot"))
				go getNewIssues(bot, update)
			case "issues":
				responses, err := getIssues()
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "There was an error while getting the issues"))
				}
				for _, response := range responses {
					if response.State == "open" {
						_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(response.Title+"\n\n"+response.Body+"\n\n"+response.Url)))
					}
				}
			}
		} else {
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid Message"))
		}
	}
}

// getIssues gets issues for our repo from curl request to github
func getIssues() ([]Response, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", User, Repo), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	var data []Response
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getNewIssues(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	time.Sleep(UpdateCycle * time.Minute)
	responses, err := getIssues()
	if err != nil {
		log.Fatal(err)
	}
	for _, response := range responses {
		if response.State == "open" && time.Now().Unix()-response.CreatedAt.Unix() <= UpdateCycle*60 {
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("NEW ISSUE!\n\n"+response.Title+"\n\n"+response.Body+"\n\n"+response.Url)))
		}
	}
}
