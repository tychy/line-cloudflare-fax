package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/fetch"
)

var channelAccessToken = os.Getenv("LINE_CHANNEL_ACCESS_TOKEN")
var channelToken = os.Getenv("LINE_CHANNEL_TOKEN")

const replyEndpoint = "https://api.line.me/v2/bot/message/reply"

type TextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ReplyMessage struct {
	ReplyToken string        `json:"replyToken"`
	Messages   []TextMessage `json:"messages"`
}

func main() {
	bot, err := linebot.New(
		channelToken,
		channelAccessToken,
	)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					replyMessage := ReplyMessage{
						ReplyToken: event.ReplyToken,
						Messages: []TextMessage{
							{
								Type: "text",
								Text: message.Text,
							},
						},
					}
					jsonData, err := json.Marshal(replyMessage)
					if err != nil {
						log.Fatalf("Error: %s", err)
						return
					}

					cli := fetch.NewClient()
					r, err := fetch.NewRequest(req.Context(), http.MethodPost, replyEndpoint, bytes.NewBuffer(jsonData))
					if err != nil {
						log.Fatal(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					r.Header.Set("Content-Type", "application/json")
					r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channelAccessToken))
					res, err := cli.Do(r)
					if err != nil {
						log.Fatal(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					defer res.Body.Close()
					body, _ := io.ReadAll(res.Body)
					fmt.Print(string(body))

				}
			}
		}
	})

	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		msg := "Hello!"
		w.Write([]byte(msg))
	})
	http.HandleFunc("/echo", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, req.Body)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		msg := "Top Page"
		w.Write([]byte(msg))
	})
	workers.Serve(nil)
}
