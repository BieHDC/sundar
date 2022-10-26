package main

import (
	"strings"
	//"maunium.net/go/mautrix"
	//event "maunium.net/go/mautrix/event"
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
		minpowerlevel := cmdhdlr.needsAtLeastPowerlevel(targetroom, commandname, cmdint.RequiredPowerlevel)
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


type HelpTextCachePowerlevel map[int][2]string
type HelpTextCache map[id.RoomID]HelpTextCachePowerlevel //room -> userpower -> helptext,helptext_formatted
func NewHelpTextCache() HelpTextCache {
	cache := make(HelpTextCache)
	return cache
}

func (cmdhdlr *CommandHandler) getHelp(htc HelpTextCache, targetroom id.RoomID, powerlevel int) (string, string) {
	cache, exists := htc[targetroom][powerlevel]
	if exists {
		return cache[0], cache[1]
	} else {
		helperstring, helperstring_formatted := cmdhdlr.generateHelp(targetroom, powerlevel)
		if htc[targetroom] == nil {
			htc[targetroom] = make(HelpTextCachePowerlevel)
		}
		htc[targetroom][powerlevel] = [2]string{helperstring, helperstring_formatted}
		return helperstring, helperstring_formatted
	}
}


func (htc HelpTextCache) invalidateHelpTextCacheForRoom(targetroom id.RoomID) {
	delete(htc, targetroom)
}

func (htc HelpTextCache) invalidateHelpTextCacheForAll() {
	htc = make(HelpTextCache)
}

func printHelp(ca CommandArgs) BotReply {
	userpower := 0
	if ca.sender == ca.cmdhdlr.botmaster {
		userpower = math.MaxInt
	} else {
		userpower = ca.cmdhdlr.masters[ca.room][ca.sender]
	}

	helperstring, helperstring_formatted := ca.cmdhdlr.getHelp(ca.cmdhdlr.helptextcache, ca.room, userpower)
	return BotPrintFormatted(ca.statusroom, helperstring, helperstring_formatted)
}


func printUsage(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		// Print its own usage
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	} else {
		// Print the requested commands usage
		return BotPrintSimple(ca.statusroom, ca.cmdhdlr.getCommandUsage(ca.argv[1]))
	}
}


func (cmdhdlr *CommandHandler) getCommandUsage(commandname string) string {
	command, ok := cmdhdlr.allcommands[commandname]
	if ok {
		return command.Usage
	} else {
		return "Command not found."
	}
}