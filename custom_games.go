package main

import (
	"math/rand"
	"strings"
	"time"
	//"maunium.net/go/mautrix"
	//event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"sync"
	"strconv"
)


func HandleGuessHandler() CommandCallback {
	type gamestate struct {
		game_active 	bool
		current_number 	int
		user_guess 		map[id.UserID]int
	}
	rooms := make(map[id.RoomID]gamestate)
	var gamemutex sync.Mutex

	return func(ca CommandArgs) BotReply {
		game := rooms[ca.room]
		defer func(){rooms[ca.room] = game}()

		gamemutex.Lock()
		defer gamemutex.Unlock()

		min, max := 1, 100
		// Init new game
		if game.game_active == false {
			rand.Seed(time.Now().UnixNano())
			game.current_number = rand.Intn(max-min) + min
			game.user_guess = make(map[id.UserID]int)
			game.game_active = true
			return BotPrintSimple(ca.room, "Guess a number between 1 and 100.\nPlease input your guess.")
		}

		// Game Running
		if ca.argc < 2 {
			return BotPrintSimple(ca.room, ca.self.Usage)
		}

		guess, err := strconv.Atoi(strings.TrimSuffix(ca.argv[1], "\n"))
		if err != nil {
			return BotPrintSimple(ca.room, "Invalid input. Please enter an integer value.")
		}

		// Only count the guess if the input was valid
		game.user_guess[ca.sender] += 1

		// Directly reference the user in the message
		if guess > game.current_number {
			return BotPrintPinged(ca.room, &ca.sender, "", "'s guess is bigger than the secret number. Try again.")
		} else if guess < game.current_number {
			return BotPrintPinged(ca.room, &ca.sender, "", "'s guess is smaller than the secret number. Try again.")
		} else {
			// Game End
			usertries := strconv.Itoa(game.user_guess[ca.sender])
			game.user_guess = nil
			game.game_active = false
			return BotPrintPinged(ca.room, &ca.sender, "Correct! ", " guessed right after " + usertries + " attempts.")
		}
	}
}