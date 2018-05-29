package main

import (
	"flag"
	"fmt"
	"guypj/lichess/api"
	"guypj/lichess/bot"
	"sync"
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
	state := bot.NewState(client)

	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	go bot.ListenForEventsForever(state, &waitGroup)
	go bot.AcceptChallengesForever(state, &waitGroup)
	go bot.PlayGamesForever(state, &waitGroup)

	waitGroup.Wait()
}
