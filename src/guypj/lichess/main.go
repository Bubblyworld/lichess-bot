package main

import (
	"flag"
	"fmt"
	"log"

	"guypj/lichess/api"
)

var apiKey = flag.String("api-key", "", "The Lichess API key to use for this bot's requests.")

func main() {
	flag.Parse()
	if *apiKey == "" {
		fmt.Println("Lichess-Bot requires a Lichess API key in order to run.")
		flag.PrintDefaults()

		return
	}

	client := api.NewLichessClient(*apiKey)
	log.Println("Lichess-Bot starting.")

	eventsChannel, err := client.StreamEvents()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	for {
		msg, ok := <-eventsChannel
		if !ok {
			break
		}

		log.Printf("Received message of type: %d", msg.Type)
	}

	log.Println("Lichess-Bot stopping.")
}
