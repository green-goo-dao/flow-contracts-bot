package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bjartek/overflow"
	"github.com/bwmarrin/discordgo"
)

func main() {

	var err error
	sleepDuration := 1 * time.Minute
	sleep := os.Getenv("SLEEP")
	if sleep != "" {
		sleepDuration, err = time.ParseDuration(sleep)
		if err != nil {
			fmt.Printf("Invalid format for SLEEP, use a valid duration %v\n", err)
			os.Exit(1)
		}
	}

	usage := `flow-contracts-bot has the following configuration options
  - DISCORD_WEBHOOK_URL: an URL from a discrod server to post messages to. (required)
	- NETWORK: the flow network to listen for events in testnet|mainnet (required)
	- EVENTS_FILE: the name of the events file to track events in, defulat:<network>.events in same dir as binary
	- SLEEP: a string duration of how long to sleep between each run, default:1m
`

	eventHookUrl, ok := os.LookupEnv("DISCORD_WEBHOOK_URL")
	if !ok {
		fmt.Println(usage)
		os.Exit(1)
	}

	network, ok := os.LookupEnv("NETWORK")
	if !ok {
		fmt.Println(usage)
		os.Exit(1)
	}

	fmt.Println(network)
	eventsFile, ok := os.LookupEnv("EVENTS_FILE")
	if !ok {
		eventsFile = fmt.Sprintf("%s.events", network)
	}

	o := overflow.Overflow(
		overflow.WithNetwork(network),
	)

	for {
		events, err := o.FetchEvents(
			overflow.WithWorkers(1),
			overflow.WithTrackProgressIn(eventsFile),
			overflow.WithEventIgnoringField("flow.AccountContractAdded", []string{"codeHash"}),
			overflow.WithEventIgnoringField("flow.AccountContractUpdated", []string{"codeHash"}),
		)
		if err != nil {
			fmt.Printf("Error fetch events %v\n", err)
			os.Exit(1)
		}

		if len(events) == 0 {
			fmt.Println("found no events")
			time.Sleep(sleepDuration)
			continue
		}

		discord, err := discordgo.New()
		if err != nil {
			fmt.Printf("Error creating discord %v\n", err)
			os.Exit(1)
		}

		parts := strings.Split(eventHookUrl, "/")
		length := len(parts)
		id := parts[length-2]
		token := parts[length-1]

		for _, event := range events {

			action := "updated"
			if event.Name == "flow.AccountContractAdded" {
				action = "added"

			}

			address := event.Event.Fields["address"].(string)

			address = strings.ReplaceAll(address, "0x", "A.")

			embed := &discordgo.MessageEmbed{
				URL:   fmt.Sprintf("https://flowscan.org/contract/%s.%s/overview", address, event.Event.Fields["contract"]),
				Title: fmt.Sprintf("%s.%s was %s", address, event.Event.Fields["contract"], action),
				Type:  discordgo.EmbedTypeLink,
				Fields: []*discordgo.MessageEmbedField{{
					Name:   "tx",
					Value:  event.Event.TransactionId,
					Inline: false,
				}, {
					Name:   "blockHeight",
					Value:  fmt.Sprintf("%d", event.BlockHeight),
					Inline: false,
				}},
			}
			message := &discordgo.WebhookParams{
				Embeds: []*discordgo.MessageEmbed{embed},
			}

			_, err = discord.WebhookExecute(id, token, true, message)

			if err != nil {
				fmt.Printf("error sending message %v\n", err)
				os.Exit(1)
			}
		}
	}
}
