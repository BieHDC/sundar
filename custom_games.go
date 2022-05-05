package main

import (
	"math/rand"
	"strings"
	"time"
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"sync"
	"strconv"
)



var game_active = false
var current_number = 0
var user_guess = make(map[id.UserID]int)
var guess_mutex sync.Mutex
// Every game is global to the bot, dont know if bug or feature yet
func HandleGuess(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	guess_mutex.Lock()
	defer guess_mutex.Unlock()

	min, max := 1, 100
	// Init new game
	if game_active == false {
		rand.Seed(time.Now().UnixNano())
		current_number = rand.Intn(max-min) + min
		BotReplyMsg(cmdhdlr, room, "Guess a number between 1 and 100.\nPlease input your guess.")
		game_active = true
		return true
	}

	// Game Running
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], room)
		return false
	}

	guess, err := strconv.Atoi(strings.TrimSuffix(argv[1], "\n"))
	if err != nil {
		BotReplyMsg(cmdhdlr, room, "Invalid input. Please enter an integer value.")
		return false
	}

	// Only count the guess if the input was valid
	user_guess[sender] += 1

	// Directly reference the user in the message
	if guess > current_number {
		BotReplyMsgPinged(cmdhdlr, sender, room, "", "'s guess is bigger than the secret number. Try again.")
	} else if guess < current_number {
		BotReplyMsgPinged(cmdhdlr, sender, room, "", "'s guess is smaller than the secret number. Try again.")
	} else {
		// Game End
		BotReplyMsgPinged(cmdhdlr, sender, room, "Correct! ", " guessed right after " + strconv.Itoa(user_guess[sender]) + " attempts.")
		for key := range user_guess {
			delete(user_guess, key)
		}
		game_active = false
	}

	return true
}