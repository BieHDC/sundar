package main

import (
	"strings"
	//"maunium.net/go/mautrix"
	//event "maunium.net/go/mautrix/event"
	//id "maunium.net/go/mautrix/id"
)


func getErrors(parsederror string) string {
	errors, exists := WinErrorCodes[parsederror]
	if !exists {
		return "Error Code not found: " + parsederror
	}

	reply := "The following codes have been found:\n"
	for _, error := range errors {
		reply += "\t" + error.Type + " -> " + error.Text + "\n"
	}
	return reply
}

func sanitiseinputandgeterrors(errorstring string) string {
	parsederrorcode := strings.TrimPrefix(errorstring, "0x")
	parsederrorcode =  strings.TrimLeft(parsederrorcode, "0")
	parsederrorcode =  strings.TrimRight(parsederrorcode, "\n")
	parsederrorcode =  strings.ToUpper(parsederrorcode)
	return getErrors(parsederrorcode)
}

func HandleErrorCodeRequest(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		return BotPrintSimple(ca.room, ca.self.Usage)
	}

	return BotPrintSimple(ca.room, sanitiseinputandgeterrors(ca.argv[1]))
}
