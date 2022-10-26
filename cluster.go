package main

import (
    //"maunium.net/go/mautrix"
    //event "maunium.net/go/mautrix/event"
    //id "maunium.net/go/mautrix/id"
    "strings"
)


func HandleRoomlistCommand(ca CommandArgs) BotReply {
    if ca.argc < 3 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	// argv 1  -> roomlist
	// argv 2  -> command
	// argv 3+ -> command args

	if ca.argv[2] == ca.argv[0] {
		return BotPrintSimple(ca.statusroom, "I will not let you recursively call this.")
	}

	roomlist := strings.Split(ca.argv[1], ",")
	if len(roomlist) < 1 {
		return BotPrintSimple(ca.statusroom, "Empty Roomlist.")
	}

	chain := BotPrintNothing()
	//build a new argc and argv
	newargv := ca.argv[2:]
	newargc := len(newargv)
	for _, targetroom := range roomlist {
		targetroomparsed, err := parseRoomID(targetroom)
		if err != nil {
			errchained := BotPrintSimple(ca.statusroom, "Failed to parse the roomid: "+err.Error())
			BotPrintAppend(&chain, &errchained)
			continue
		}
		success := InvokeCommand(ca.cmdhdlr, targetroomparsed.RoomID(), ca.sender, newargc, newargv, ca.statusroom, ca.evt)
		if !success {
			return BotPrintSimple(ca.statusroom, "An error occurred while executing the command, stopping.")
		}
	}

	chainfin := BotPrintSimple(ca.statusroom, "Command executed.")
	BotPrintAppend(&chain, &chainfin)
	return chain
}


// Goodie for Cluster
func HandleBotSayMessage(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	return BotPrintPinged(ca.room, &ca.sender, "", " says: " + strings.Join(ca.argv[1:], " "))
}