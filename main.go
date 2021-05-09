package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	selectVersionAction = "select-version"
)

func main() {
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	http.HandleFunc("/slack/events", slackVerificationMiddleware(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
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
				case "music":
					text := slack.NewTextBlockObject(slack.MarkdownType, "*Please choose how you feel now* :ghost:", false, false)
					textSection := slack.NewSectionBlock(text, nil, nil)

					feelings := []string{"chill", "rock", "nogizaka46"}
					options := make([]*slack.OptionBlockObject, 0, len(feelings))
					var headphones string = ":headphones:"
					headphonesObj := slack.NewTextBlockObject(slack.PlainTextType, headphones, false, false)
					for _, v := range feelings {
						optionText := slack.NewTextBlockObject(slack.PlainTextType, v, false, false)
						options = append(options, slack.NewOptionBlockObject(v, headphonesObj, optionText))
					}

					placeholder := slack.NewTextBlockObject(slack.PlainTextType, "how you feel now", false, false)
					selectMenu := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, placeholder, "", options...)

					actionBlock := slack.NewActionBlock(selectVersionAction, selectMenu)

					fallbackText := slack.MsgOptionText("This client is not supported.", false)
					blocks := slack.MsgOptionBlocks(textSection, actionBlock)

					if _, err := api.PostEphemeral(event.Channel, event.User, fallbackText, blocks); err != nil {
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
								Text: "Please type *music* or *ping*",
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
	}))

	http.HandleFunc("/slack/actions", slackVerificationMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var payload *slack.InteractionCallback
		if err := json.Unmarshal([]byte(r.FormValue("payload")), &payload); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch payload.Type {
		case slack.InteractionTypeBlockActions:
			if len(payload.ActionCallback.BlockActions) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			action := payload.ActionCallback.BlockActions[0]
			switch action.BlockID {
			case selectVersionAction:
				selectVal := action.SelectedOption.Value

				deleteOriginal := slack.MsgOptionDeleteOriginal(payload.ResponseURL)
				if _, _, _, err := api.SendMessage("", deleteOriginal); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				t := time.Now()

				switch selectVal {
				case "chill":
					switch t.Second() % 3 {
					case 0:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=fCbYoBTIBDU&list=PLN5VowuWXMKbRxDAu3EiftGIRyDR4Ar9v&index=23|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					case 1:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=DhcIUYHiJDI|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					case 2:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=ZeIGVnkYX04|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					}
				case "rock":
					switch t.Second() % 3 {
					case 0:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=bs1rRIVBzh0&list=PLN5VowuWXMKaXddkh5NkCRGdoW3mXsCuq&index=10|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					case 1:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=yXq_KSBmpEI&list=PLN5VowuWXMKaXddkh5NkCRGdoW3mXsCuq&index=13|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					case 2:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=FAU7NyiqwQY|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					}
				case "nogizaka46":
					switch t.Second() % 3 {
					case 0:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=jEZv7KWhVrU&list=PLUae9wntmJNMvIRkoHL98S462nqZshKmW&index=75|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					case 1:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=NvitdI9IGTM&list=PLUae9wntmJNMvIRkoHL98S462nqZshKmW&index=124|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					case 2:
						msg := slack.MsgOptionText(
							fmt.Sprintf(":headphones: *%s* <https://www.youtube.com/watch?v=m-DngT-YioA&list=PLUae9wntmJNMvIRkoHL98S462nqZshKmW&index=40|Recommended songs>", selectVal), false)
						if _, _, err := api.PostMessage(payload.Channel.ID, msg); err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}))

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func slackVerificationMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(w, r)
	}
}
