package main

import (
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	gomoji "github.com/forPelevin/gomoji"
)

func snarkyReplies(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) {
	if argv[0] == "sudo" {
		command, ok := cmdhdlr.allcommands[argv[1]]
		if ok {
			if !cmdhdlr.hasUserRequiredPowerlevel(room, sender, argv[1], &command) {
				msg_1 := "sudo: User "
				msg_2 := " is not in the sudoers file. The Sysadmin will now hunt you down."
				BotReplyMsgPinged(cmdhdlr, sender, room, msg_1, msg_2)
			} else {
				BotReplyMsg(cmdhdlr, room, "sudo: The command you tried to dial is current not available.")
			}
		} else {
			BotReplyMsg(cmdhdlr, room, "sudo: command not found.")
		}
	} else if argv[0] == "doas" {
		command, ok := cmdhdlr.allcommands[argv[1]]
		if ok {
			if !cmdhdlr.hasUserRequiredPowerlevel(room, sender, argv[1], &command) {
				msg_1 := "doas: User "
				msg_2 := " is based, but still not permitted to run this command."
				BotReplyMsgPinged(cmdhdlr, sender, room, msg_1, msg_2)
			} else {
				BotReplyMsg(cmdhdlr, room, "doas: Yes, i actually invested time into making those.")
			}
		} else {
			BotReplyMsg(cmdhdlr, room, "Based, but command not found")
		}
	}
}

func emojireactor(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) {
	// Finds all emojis in a message and adds them as reaction
	for i := 0; i < argc; i++ {
		for _, element := range gomoji.FindAll(argv[i]) {
	    	cmdhdlr.client.SendReaction(room, evt.ID, element.Character)
	    }
	}
}

func goodoldfriend(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) {
	// User specific trolls, add a certain reaction if this bot sees a message from og dev, me
	if sender == id.UserID("@biehdc:reactos.org") {
		cmdhdlr.client.SendReaction(room, evt.ID, "🐸") // A frog
		cmdhdlr.client.SendReaction(room, evt.ID, "🇦🇹") // An austrian flag
		return
	}
}

