package main

import (
	//event "maunium.net/go/mautrix/event"
	//id "maunium.net/go/mautrix/id"
	"os"
	"encoding/json"
	"strings"
)


type EchoRegister struct {
	echofile 		string
	Allechos 		map[string]string //Key: Callsign; Value: The message (string)
}

func NewEchoRegister(echofile string) *EchoRegister {
    return &EchoRegister{
    	echofile: echofile,
    	Allechos: make(map[string]string),
    }
}

func HandleBotEchoMessage(ca CommandArgs) BotReply {
	text, exists := ca.cmdhdlr.echoregister.Allechos[ca.argv[0]]
	if !exists {
		return BotPrintSimple(ca.statusroom, "Echo called >" + ca.argv[0] + "< does not exist.")
	}

	return BotPrintSimple(ca.room, text)
}



type Echo struct {
	Callsign string
	Powerlevel int
	Message string
}

/*
func EchoAdd(cmdhdlr *CommandHandler, callsign string, powerlevel int, message string, category string) {
	emitCommand(cmdhdlr, Echo{Callsign: callsign, Powerlevel: powerlevel, Message: message}, category)
}
*/

func loadEchos(cmdhdlr *CommandHandler, er *EchoRegister) {
	if er.echofile == "NONE" {
		botNotifyEventsChannel(cmdhdlr, "No echo file specified, skipping")
		return
	}
	data, err := os.ReadFile(er.echofile)
	if err != nil {
		panic(err)
	}
	var echos []Echo
	err = json.Unmarshal(data, &echos)
	if err != nil {
		panic(err)
	}

    for _, echo := range echos {
		emitCommand(cmdhdlr, er, echo, "Echo")
    }
}

func emitCommand(cmdhdlr *CommandHandler, er *EchoRegister, echo Echo, category string) bool {
	var desc string

	fixedmessage := strings.ReplaceAll(echo.Message, "\n", " ")

    if len(fixedmessage) <= 70 {
        desc = fixedmessage
    } else {
        desc = fixedmessage[:67]+"..."
    }

	success := cmdhdlr.AddCommandChecked(echo.Callsign, desc, "", category, echo.Powerlevel, HandleBotEchoMessage)
	if success {
		er.Allechos[echo.Callsign] = echo.Message
	}

	return success
}