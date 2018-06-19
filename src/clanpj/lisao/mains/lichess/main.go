package main

import (
	"flag"
	"fmt"
	"sync"

	"clanpj/lisao/lichess"
)

var apiKey = flag.String("api-key", "", "The Lichess API key to use for this bot's requests.")

func main() {
	flag.Parse()
	if *apiKey == "" {
		fmt.Println("Lichess-Bot requires a Lichess API key in order to run.")
		flag.PrintDefaults()

		return
	}

	client := lichess.NewLichessClient(*apiKey)
	state := NewState(client)

	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	go ListenForEventsForever(state, &waitGroup)
	go AcceptChallengesForever(state, &waitGroup)
	go PlayGamesForever(state, &waitGroup)

	waitGroup.Wait()
}
