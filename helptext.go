package main

import (
	"strings"
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"math"
	"sort"
	"strconv"
	html "golang.org/x/net/html"
)

func helptextToCodeblock(unformatted string) string {
	stringbuilder := strings.TrimPrefix(unformatted, "```\n")      //remove the initiator
	stringbuilder = strings.TrimSuffix(stringbuilder, "\n```")     //remove the close clause
	//stringbuilder = strings.ReplaceAll(stringbuilder, ">", "&gt;") //html fun
	stringbuilder = html.EscapeString(stringbuilder) 				 //better html fun
	stringbuilder = "<pre><code>" + stringbuilder + "</code></pre>\n"
	return stringbuilder
}


func (cmdhdlr *CommandHandler) generateHelp(targetroom id.RoomID, powerlevel int) (string, string) {
	types := make(map[string][]string)
	for commandname, cmdint := range cmdhdlr.allcommands {
		minpowerlevel := cmdhdlr.needsAtLeastPowerlevel(targetroom, commandname, &cmdint)
		if powerlevel >= minpowerlevel {
			appendee := "\t" + commandname + " -> " + cmdint.Description

			if minpowerlevel > CommandAnyone {
				appendee += " -> Required Power Level: " + strconv.Itoa(minpowerlevel)
			}
			appendee += "\n"
			types[cmdint.Type_] = append(types[cmdint.Type_], appendee)
		}
	}

	keys := make([]string, 0, len(types))
	for key := range types {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	helperstring := "```\nAvailable Commands for you in this room:\n"
	for _, key := range keys {
		helperstring += key + ":\n"

		entries := make([]string, 0, len(types[key]))
		for entry := range types[key] {
			entries = append(entries, types[key][entry])
		}
		sort.Strings(entries)

		for _, index := range entries {
			helperstring += index
		}
		helperstring += "\n"
	}
	helperstring += "```"

	return helperstring, helptextToCodeblock(helperstring)
}


var helptextcache map[id.RoomID]map[int][2]string //room -> userpower -> helptext,helptext_formatted
func (cmdhdlr *CommandHandler) getHelp(targetroom id.RoomID, powerlevel int) (string, string) {
	if helptextcache == nil {
		helptextcache = make(map[id.RoomID]map[int][2]string)
	}

	cache, exists := helptextcache[targetroom][powerlevel]
	if exists {
		return cache[0], cache[1]
	} else {
		helperstring, helperstring_formatted := cmdhdlr.generateHelp(targetroom, powerlevel)
		if helptextcache[targetroom] == nil {
			helptextcache[targetroom] = make(map[int][2]string)
		}
		helptextcache[targetroom][powerlevel] = [2]string{helperstring, helperstring_formatted}
		return helperstring, helperstring_formatted
	}
}


func (cmdhdlr *CommandHandler) invalidateHelpTextCacheForRoom(targetroom id.RoomID) {
	delete(helptextcache, targetroom)
}

func (cmdhdlr *CommandHandler) invalidateHelpTextCacheForAll() {
	helptextcache = make(map[id.RoomID]map[int][2]string)
}

func printHelp(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	userpower := 0
	if sender == cmdhdlr.botmaster {
		userpower = math.MaxInt
	} else {
		userpower = cmdhdlr.masters[room][sender]
	}

	helperstring, helperstring_formatted := cmdhdlr.getHelp(room, userpower)
	BotReplyMsgFormatted(cmdhdlr, statusroom, helperstring, helperstring_formatted)
	return true
}


func printUsage(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 2 {
		// Print its own usage
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
	} else {
		// Print the requested commands usage
		cmdhdlr.internelPrintUsage(argv[1], statusroom)
	}
	return true
}


func (cmdhdlr *CommandHandler) internelPrintUsage(commandname string, roomid id.RoomID) {
	command, ok := cmdhdlr.allcommands[commandname]
	if ok {
		BotReplyMsg(cmdhdlr, roomid, command.Usage)
	} else {
		BotReplyMsg(cmdhdlr, roomid, "Command not found.")
	}
}