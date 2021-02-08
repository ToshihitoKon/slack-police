package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var (
	token           = os.Getenv("SLAKC_POLICE_TOKEN")
	signingSecret   = os.Getenv("SLACK_POLICE_SIGNING_SECRET")
	policeChannelID = os.Getenv("SLACK_POLICE_EMOJI_CHANNNEL_ID")
	port            = 5002
	api             = slack.New(token)
)

func main() {

	http.HandleFunc("/events-endpoint", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			log.Println("recieved: ", innerEvent.Type)

			switch ev := innerEvent.Data.(type) {
			case *slackevents.EmojiChangedEvent:
				switch ev.Subtype {
				case "add":
					log.Println(
						api.PostMessage(
							policeChannelID,
							slack.MsgOptionText(
								fmt.Sprintf("絵文字警察です👮\n:%s: %s が追加されました", ev.Name, ev.Name),
								false)))
				case "remove":
					log.Println(
						api.PostMessage(
							policeChannelID,
							slack.MsgOptionText(
								fmt.Sprintf("絵文字警察です👮\n%s が消えました", ev.Names[0]),
								false)))
				case "rename":
					log.Println(
						api.PostMessage(
							policeChannelID,
							slack.MsgOptionText(
								fmt.Sprintf("絵文字警察です👮\n:%s: の名前がかわりました\n%s -> %s", ev.NewName, ev.OldName, ev.NewName),
								false)))
				default:
					log.Println("dismiss: ", ev)
				}
			}
		}
	})
	fmt.Println("[INFO] slack-police")
	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
