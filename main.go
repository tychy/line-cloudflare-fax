package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare/fetch"
)

type TextMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ReplyMessage struct {
	ReplyToken string        `json:"replyToken"`
	Messages   []TextMessage `json:"messages"`
}

func genReplyText(replyToken string, text string) ([]byte, error) {
	replyMessage := ReplyMessage{
		ReplyToken: replyToken,
		Messages: []TextMessage{
			{
				Type: "text",
				Text: text,
			},
		},
	}
	jsonData, err := json.Marshal(replyMessage)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func getContent(ctx context.Context, id string) error {
	log.Print("getContent", id)
	url := fmt.Sprintf(contentsEndpoint, id)
	cli := fetch.NewClient()
	r, err := fetch.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channelAccessToken))

	log.Print("execute request", url)
	res, err := cli.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	fmt.Print(string(body))
	return nil
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
				var jsonData []byte
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					jsonData, err = genReplyText(event.ReplyToken, message.Text)
					if err != nil {
						log.Fatalf("Error: %s", err)

						w.WriteHeader(http.StatusInternalServerError)
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

				case *linebot.FileMessage:
					id := message.ID
					err := getContent(req.Context(), id)
					if err != nil {
						log.Fatalf("Error: %s", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					jsonData, err = genReplyText(event.ReplyToken, "uploaded")
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
				default:
					continue
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
