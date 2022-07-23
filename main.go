package main

import (
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func mustEnv(name string) (string, bool) {
	value := os.Getenv(name)
	if value == "" {
		log.Println("must env ", name)
		return "", false
	}
	return value, true
}

func main() {
	isValid := true

	botToken, ok := mustEnv("SLACK_POLICE_BOT_TOKEN")
	if !ok {
		isValid = false
	}
	appToken, ok := mustEnv("SLACK_POLICE_APP_TOKEN")
	if !ok {
		isValid = false
	}
	policeChannelID, ok := mustEnv("SLACK_POLICE_EMOJI_CHANNEL_ID")
	if !ok {
		isValid = false
	}
	if !isValid {
		return
	}

	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)
	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)
	go runner(api, client, policeChannelID)

	fmt.Println("[INFO] slack-police")
	fmt.Println("[INFO] run websocket")
	client.Run()

}

func runner(api *slack.Client, client *socketmode.Client, policeChannelID string) {
	for evt := range client.Events {
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			fmt.Println("Connecting to Slack with Socket Mode...")
		case socketmode.EventTypeConnectionError:
			fmt.Println("Connection failed. Retrying later...")
		case socketmode.EventTypeConnected:
			fmt.Println("Connected to Slack with Socket Mode.")
		case socketmode.EventTypeEventsAPI:
			eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
			if !ok {
				fmt.Printf("Ignored %+v\n", evt)
				continue
			}
			client.Ack(*evt.Request)

			switch eventsAPIEvent.Type {
			case slackevents.CallbackEvent:
				procInnerEvent(api, eventsAPIEvent.InnerEvent, policeChannelID)
			}
		}
	}
}

func procInnerEvent(api *slack.Client, event slackevents.EventsAPIInnerEvent, policeChannelID string) {
	log.Println("recieved: ", event.Type)

	switch ev := event.Data.(type) {
	case *slackevents.EmojiChangedEvent:
		switch ev.Subtype {
		case "add":
			log.Println(
				api.PostMessage(
					policeChannelID,
					slack.MsgOptionText(
						fmt.Sprintf("絵文字警察です👮\n:%s: %s が追加されました", ev.Name, ev.Name),
						false,
					),
				),
			)
		case "remove":
			log.Println(
				api.PostMessage(
					policeChannelID,
					slack.MsgOptionText(
						fmt.Sprintf("絵文字警察です👮\n:%s: が消えました", ev.Names[0]),
						false),
				),
			)
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
