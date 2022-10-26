package main

import (
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	gomoji "github.com/forPelevin/gomoji"
)

func snarkyRepliesHandler() FunctionRegisterPrototype {
	type troll struct {
		//match string //key
		msg_1 string
		msg_2 string
		plot string
		not_found string
	}
	trolls := map[string]troll{
		"sudo": troll{	msg_1: "sudo: User ", 
						msg_2: " is not in the sudoers file. The Sysadmin will now hunt you down.", 
						plot: "sudo: The command you tried to dial is current not available.",
						not_found: "sudo: command not found."},

		"doas": troll{	msg_1: "doas: User ", 
						msg_2: " is based, but still not permitted to run this command.", 
						plot: "doas: Yes, i actually invested time into making those.",
						not_found: "Based, but command not found"},
	}

	return func(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) BotReply {
		whichone, found := trolls[argv[0]]
		if !found {
			return BotPrintNothing()
		}

		command, ok := cmdhdlr.allcommands[argv[1]]
		if !ok {
			return BotPrintSimple(room, whichone.not_found)
		}

		if !cmdhdlr.hasUserRequiredPowerlevel(room, sender, argv[1], command.RequiredPowerlevel) {
			return BotPrintPinged(room, &sender, whichone.msg_1, whichone.msg_2)
		}

		return BotPrintSimple(room, whichone.plot)
	}
}

func emojireactor(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) BotReply {
	// Finds all emojis in a message and adds them as reaction
	for i := 0; i < argc; i++ {
		for _, element := range gomoji.FindAll(argv[i]) {
	    	cmdhdlr.client.SendReaction(room, evt.ID, element.Character)
	    }
	}
	return BotPrintNothing()
}

func goodoldfriend(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) BotReply {
	// User specific trolls, add a certain reaction if this bot sees a message from og dev, me
	if sender == id.UserID("@biehdc:reactos.org") {
		cmdhdlr.client.SendReaction(room, evt.ID, "🐸") // A frog
		cmdhdlr.client.SendReaction(room, evt.ID, "🇦🇹") // An austrian flag
		return BotPrintNothing()
	}
	return BotPrintNothing()
}

