package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, os.Getenv("SLACK_SIGNING_SECRET"))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bodyReader := io.TeeReader(r.Body, &verifier)
		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := verifier.Ensure(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case slackevents.CallbackEvent:
			innerEvent := eventsAPIEvent.InnerEvent
			switch event := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				message := strings.Split(event.Text, " ")
				if len(message) < 2 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				command := message[1]
				options := message[2:]

				switch command {
				case "ping":
					if len(options) > 0 {
						switch {
						case len(options) > 1:
							if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionBlocks(
								&slack.SectionBlock{
									Type: slack.MBTSection,
									Text: &slack.TextBlockObject{
										Type: "mrkdwn",
										Text: "*いやちょっと多いぃ！*",
									},
								},
							)); err != nil {
								log.Println(err)
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
						case options[0] == "pong":
							if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionBlocks(
								&slack.SectionBlock{
									Type: slack.MBTSection,
									Text: &slack.TextBlockObject{
										Type: "mrkdwn",
										Text: "*いやそれ俺が言うやつぅ！*",
									},
								},
							)); err != nil {
								log.Println(err)
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
						case options[0] == "ping":
							if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionBlocks(
								&slack.SectionBlock{
									Type: slack.MBTSection,
									Text: &slack.TextBlockObject{
										Type: "mrkdwn",
										Text: "*いや同じこと2回ぃぃ‍️！*",
									},
								},
							)); err != nil {
								log.Println(err)
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
						default:
							if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionBlocks(
								&slack.SectionBlock{
									Type: slack.MBTSection,
									Text: &slack.TextBlockObject{
										Type: "mrkdwn",
										Text: "*いや想定外ぃぃ！*",
									},
								},
							)); err != nil {
								log.Println(err)
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
						}
					} else {
						if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("pong", false)); err != nil {
							log.Println(err)
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
					}
				default:
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionBlocks(
						&slack.SectionBlock{
							Type: slack.MBTSection,
							Text: &slack.TextBlockObject{
								Type: "mrkdwn",
								Text: "<https://www.youtube.com/watch?v=bs1rRIVBzh0&list=PLN5VowuWXMKaXddkh5NkCRGdoW3mXsCuq&index=10|my favorite songs>:ghost:",
							},
						},
					)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}
		}
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
