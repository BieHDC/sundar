package main

import (
	//"fmt"
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	//"strconv"
	"strings"
	"sync"
)


type EchoInternal struct {
    Message 	string 		// The Message
    Admin 		id.UserID 	// The Owner of the Message
}

type EchoRegister struct {
	Allechos 		map[string]EchoInternal //Key: Callsign; Value EchoInternalStruct
	echo_mutex 		sync.Mutex
}

func NewEchoRegister() *EchoRegister {
    return &EchoRegister{
    	Allechos: make(map[string]EchoInternal),
    }
}

//AccountStorage will take care of caching
func saveEchos(cmdhdlr *CommandHandler, ecreg *EchoRegister) {
    cmdhdlr.StoreData("echoregister_1", ecreg)
}

func loadEchos(cmdhdlr *CommandHandler, ecreg *EchoRegister) {
    cmdhdlr.FetchData("echoregister_1", ecreg)
    for callsign, echoinfo := range ecreg.Allechos {
    	emitCommand(cmdhdlr, callsign, echoinfo.Message)
    }
	cmdhdlr.invalidateHelpTextCacheForAll()
}

func emitCommand(cmdhdlr *CommandHandler, callsign string, msg string) bool {
	desc := ""
    if len(msg) <= 70 {
        desc = msg
    } else {
        desc = msg[:70]+"..."
    }

	if !cmdhdlr.AddCommandChecked(callsign, desc, "", "Echo", CommandAnyone, HandleBotFormattedMessage) {
		return false
	} else {
		return true
	}
}


func HandleBotFormattedMessageAdd(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	cmdhdlr.echoregister.echo_mutex.Lock()
	defer cmdhdlr.echoregister.echo_mutex.Unlock()

	if argc < 3 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		msg := strings.Join(argv[2:], " ")
		if !emitCommand(cmdhdlr, argv[1], msg) {
			BotReplyMsg(cmdhdlr, statusroom, "A command with the callsign >" + argv[1] + " already exists!")
			return false
		}
		cmdhdlr.invalidateHelpTextCacheForAll()
		cmdhdlr.echoregister.Allechos[argv[1]] = EchoInternal{Message: msg, Admin: sender}
		saveEchos(cmdhdlr, cmdhdlr.echoregister)

		BotReplyMsg(cmdhdlr, statusroom, "Echocommand >" + argv[1] + "< has been successfully added.")
	}
	return true
}

func HandleBotFormattedMessageRemove(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	cmdhdlr.echoregister.echo_mutex.Lock()
	defer cmdhdlr.echoregister.echo_mutex.Unlock()
	
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		echoinfo, exists := cmdhdlr.echoregister.Allechos[argv[1]]
		if !exists {
			BotReplyMsg(cmdhdlr, statusroom, "An Echocommand called >" + argv[1] + "< does not exist.")
			return false
		}

		if echoinfo.Admin != sender || sender != cmdhdlr.botmaster {
			BotReplyMsg(cmdhdlr, statusroom, "You are not the owner of >" + argv[1] + "<.")
			return false
		}

		cmdhdlr.removeCommand(argv[1])
		delete(cmdhdlr.echoregister.Allechos, argv[1])
		saveEchos(cmdhdlr, cmdhdlr.echoregister)

		BotReplyMsg(cmdhdlr, statusroom, "Echocommand >" + argv[1] + "< has been successfully removed.")
	}
	return true
}


func HandleBotFormattedMessage(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	cmdhdlr.echoregister.echo_mutex.Lock() // It could be removed while its being printed
	defer cmdhdlr.echoregister.echo_mutex.Unlock()
	
	text, exists := cmdhdlr.echoregister.Allechos[argv[0]]
	if !exists {
		BotReplyMsg(cmdhdlr, statusroom, "An Echocommand called >" + argv[1] + "< does not exist.")
		return false
	}

	BotReplyMsg(cmdhdlr, room, text.Message)
	return true
}