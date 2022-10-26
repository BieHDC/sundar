//todo:

//	bring back the cursed starboard? yes, eventually
//	scan for :oof: patterns and post the emoji (needs custom emoji pics support from matrix)
//	add option for the botmaster to set status (needs matrix side support)
//  tie up e2ee support once proper support for bot-e2ee is implemented


package main

import (
	//"fmt"
	"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strconv"
	"strings"
	"time"
)

// You can also provide a number directly btw
const CommandAnyone 	int = 0
const CommandModerator 	int = 50
const CommandAdmin 		int = 100
const CommandMaster 	int = 1000

//We still pass it in for a small specific amount of calls, but have it friendlier to use with if we extrace the sender and room
type CommandArgs struct {
	cmdhdlr 	*CommandHandler
	room 		id.RoomID
	sender 		id.UserID
	argc 		int
	argv 		[]string
	statusroom 	id.RoomID
	evt 		*event.Event
	self 		commandsInternal
}
type CommandCallback func(CommandArgs) BotReply
type commandsInternal struct {
	Type_              			string          	// What kind of command it is. For autohelp.
	Description        			string          	// What it does. For autohelp.
	Usage 						string 				// For the "usage" command to be picked up.
	RequiredPowerlevel 			int             	// The minimum powerlevel a user needs for the command.
	Targetfunc         			CommandCallback 	// The function to call if we match.
}

type Ratelimit struct {
	lastcall time.Time
	warningmsg id.EventID
}
const granulatymsg = time.Millisecond*100
const timelimit = time.Second * 3


type CommandHandler struct {
	client         *mautrix.Client
	eventsChannel  id.RoomID
	eventsToStdout bool
	prefix         string

	joined_rooms []id.RoomID
	accountstore *AccountStorage

	funcregister 	*FunctionRegister
	echoregister	*EchoRegister

	allcommands 	map[string]commandsInternal
	alloverrides 	map[string]map[id.RoomID]int // command -> room/powerlevel // If the command has a different powerlevel in a room

	helptextcache 	HelpTextCache

	ratelimit		map[id.UserID]Ratelimit

	botmaster  	id.UserID // Untouchable Admin, Botmaster
	masters 	map[id.RoomID]map[id.UserID]int
}

func NewCommandHandler(client *mautrix.Client, prefix string, master id.UserID, eventsChannel id.RoomID, eventsToStdout bool, accountstoreprefix string, accountstoredebug bool, echofile string) *CommandHandler {
	cmdhdlr := CommandHandler{
		client:         client,
		eventsChannel:  eventsChannel,
		eventsToStdout: eventsToStdout,
		prefix:         prefix,

		allcommands: 	make(map[string]commandsInternal),
		alloverrides: 	make(map[string]map[id.RoomID]int),
		accountstore: 	NewDefaultAccountStorage(accountstoreprefix, accountstoredebug),

		funcregister: 		NewDefaultFunctionRegister(),
		echoregister:		NewEchoRegister(echofile),

		helptextcache: 	NewHelpTextCache(),

		ratelimit:	make(map[id.UserID]Ratelimit),

		botmaster: 	master,
		masters: 	make(map[id.RoomID]map[id.UserID]int),
	}

	cmdhdlr.loadRoomsAndPowerlevelsFromUserdata()

	return &cmdhdlr
}


func (cmdhdlr *CommandHandler) Done() {
	// "Internal" Commands
	// Maintinance
	cmdhdlr.AddCommand("forceleave", 		"Force Bot to leave a room", 				"<this>OR<RoomID>", "Bot Administration", CommandAdmin, HandleBotForceLeaveRoom)
	cmdhdlr.AddCommand("forcejoin", 		"Force Bot to join a room", 				"<RoomID>", 		"Bot Administration", CommandMaster, HandleBotForceJoinRoom)
	cmdhdlr.AddCommand("leaveemptyrooms", 	"Make bot account leave all empty rooms", 	"", 				"Bot Administration", CommandMaster, HandleBotLeaveEmptyRooms)

	// Tweaking
	cmdhdlr.AddCommand("setdisplayname", 		"Change the bots displayname", 	"<New Display Name>", "Appearance", CommandMaster, HandleSetDisplayName)
	cmdhdlr.AddCommand("setavatar", 			"Change the bots avatar", 		"<mxc uri>OR<clear>", "Appearance", CommandMaster, HandleSetAvatar)

	// Power Level Management
	cmdhdlr.AddCommand("poweroverride", 		"Override the required powerlevel for a command for this room", "[roomid] <command> <new power level>", "Powermanagement", CommandAdmin, HandleCommandPowerlevelOverride)
	cmdhdlr.AddCommand("poweroverridereset", 	"Reset the override powerlevel for a command for this room", 	"[roomid] <command>", 					"Powermanagement", CommandAdmin, HandleCommandPowerlevelOverrideReset)
	cmdhdlr.AddCommand("poweroverrideresetall", "Reset all overrides for a room", 								"[roomid]",								"Powermanagement", CommandAdmin, HandleCommandPowerlevelOverrideResetForAllCommandsInRoom)
	cmdhdlr.AddCommand("blockcommand", 			"Quicker way to block a command for all but admins", 			"[roomid] <command>", 					"Powermanagement", CommandAdmin, HandleCommandPowerlevelBlock)

	// Room List
	cmdhdlr.AddCommand("broadcast", 		"Broadcast Message to all joined rooms", 	"<Your Message>", 											"Tools", CommandMaster, HandleBotBroadcastMessage)
	cmdhdlr.AddCommand("roomlistcommand",	"Run a command on a list of rooms",			"<roomlist comma separated> <command> [command arguments]",	"Tools", CommandAdmin, HandleRoomlistCommand)
	cmdhdlr.AddCommand("say", 				"Say a message in a room", 					"<Your Message>", 											"Tools", CommandAdmin, HandleBotSayMessage)

	// Functionmanagement
	cmdhdlr.AddCommand("funcstate",		"Set the State of a Function in a Room", 	"<functionid> <enable>OR<disable> <this>OR<RoomID>", 	"Functionmanagement", CommandAdmin, 	HandleFunctionStatusInRoom)
	cmdhdlr.AddCommand("funcwiperooms",	"Deletes the Roomlist for a Function", 		"<functionid>", 										"Functionmanagement", CommandMaster, 	HandleFunctionWipeRoomlist)
	cmdhdlr.AddCommand("funclist",		"List all registered Functions", 			"", 													"Functionmanagement", CommandAdmin, 	HandleListKnownFunctions)

	// Help
	cmdhdlr.AddCommand("help", 	"Displays this help message", 		"", "Help", CommandAnyone, printHelp)
	cmdhdlr.AddCommand("usage",	"Displays the usage of a command", 	"<command>\n<arg> is a required argument\n[arg] is an optional argument\nThe OR keyword tells you to choose between 2 or more options on a field.", "Help", CommandAnyone, printUsage)
	// any commands need to be added before this line


	// Load Echo Commands
	loadEchos(cmdhdlr, cmdhdlr.echoregister)

	//load power level overrides for commands
	cmdhdlr.loadPowerLevelsOverrides()

	//load function blacklist
	loadFunctionstatus(cmdhdlr, cmdhdlr.funcregister)
}

func (cmdhdlr *CommandHandler) AddCommand(targetcommand string, desc string, usage string, type_ string, powerlevel int, targetfunc CommandCallback) {
	cmdhdlr.allcommands[targetcommand] = commandsInternal{
		Description:        desc,
		Usage: 				"Usage: "+cmdhdlr.prefix+targetcommand+" "+usage,
		Type_:              type_,
		RequiredPowerlevel: powerlevel,
		Targetfunc:         targetfunc,
	}

	cmdhdlr.alloverrides[targetcommand] = make(map[id.RoomID]int)
}

func (cmdhdlr *CommandHandler) AddCommandChecked(targetcommand string, desc string, usage string, type_ string, powerlevel int, targetfunc CommandCallback) bool {
	_, exists := cmdhdlr.allcommands[targetcommand]
	if !exists {
		cmdhdlr.AddCommand(targetcommand, desc, usage, type_, powerlevel, targetfunc)
		return true
	} else {
		botNotifyEventsChannel(cmdhdlr, "The command with the callsign >" + targetcommand + "< exists already!!")
		return false
	}
}

/*
func (cmdhdlr *CommandHandler) removeCommand(targetcommand string) {
	delete(cmdhdlr.allcommands, targetcommand)
	cmdhdlr.invalidateHelpTextCacheForAll()
}
*/


func InvokeCommand(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	//fmt.Println(argv)
	command, ok := cmdhdlr.allcommands[argv[0]]
	if !ok {
		cmdhdlr.BotPrint(BotPrintSimple(statusroom, "I do not know about this command. Did you check >"+cmdhdlr.prefix+"help<?"))
		return false // If the command does not exist however, we should break the loop
	}

	if !cmdhdlr.hasUserRequiredPowerlevel(room, sender, argv[0], command.RequiredPowerlevel) {
		complaint := "HandleDispatch: User >" + sender.String() + 
						"< tried to do the admin command >" + argv[0] + 
						"< which requires at least power level " + 
						strconv.Itoa(cmdhdlr.needsAtLeastPowerlevel(room, argv[0], command.RequiredPowerlevel))
		botNotifyEventsChannel(cmdhdlr, complaint)
		cmdhdlr.BotPrint(BotPrintPinged(statusroom, &sender, "",
			": You do not have permission to run this command, this incident will be reported"))
		return true // User might have the perms to run it in the other rooms in the clusterlist
	}

	ca := CommandArgs{cmdhdlr, evt.RoomID, evt.Sender, argc, argv, evt.RoomID, evt, command}
	cmdhdlr.BotPrint(command.Targetfunc(ca))
	return true
}


// Helper for Command Parsing
func sanitiseParseCommandWithArguments(source string, prefix string) ([]string, int) {
	//trim prefix
	parsed := strings.TrimPrefix(source, prefix)
	//trim newline+spaces both ends
	parsed = strings.TrimSpace(parsed)
	//string split by spaces
	splitted := strings.Split(parsed, " ")
	//remove empty strings (can happen if there are multiple spaces between 2 arguments)
	len := 0
	for _, value := range splitted {
		if value != "" { // filter
			splitted[len] = value
			len++
		}
	}
	splitted = splitted[:len]
	// done
	return splitted, len
}

func (cmdhdlr *CommandHandler) HandleDispatch(evt *event.Event) {
	argv, argc := sanitiseParseCommandWithArguments(evt.Content.AsMessage().Body, cmdhdlr.prefix)
	if argc < 1 {
		return	
	}

	if strings.HasPrefix(evt.Content.AsMessage().Body, cmdhdlr.prefix) {
		// Simple Ratelimiter
		rl, exists := cmdhdlr.ratelimit[evt.Sender]
		if exists && rl.lastcall.Add(timelimit).Truncate(granulatymsg).After(time.Now()) {
			if rl.warningmsg == id.EventID("0") {
				warnreply := "Cool it! Wait " + timelimit.String() + "."
				warningmsg := botReplyMsg(cmdhdlr, evt.RoomID, warnreply)
				cmdhdlr.ratelimit[evt.Sender] = Ratelimit{lastcall: time.Now(), warningmsg: warningmsg}
			} else {
				timeleft := (((timelimit)-time.Now().Sub(rl.lastcall))/2).Truncate(granulatymsg)
				newtime := time.Now().Add(timeleft) //we do a little trolling
				cmdhdlr.ratelimit[evt.Sender] = Ratelimit{lastcall: newtime, warningmsg: rl.warningmsg}
				warnreply := "Cool it more! Wait for " + timeleft.String() + "."
				botReplyMsgEdited(cmdhdlr, evt.RoomID, warnreply, rl.warningmsg)
				//fixme here we could add the function for the bot to delete the spammed message
			}
		} else {
			cmdhdlr.ratelimit[evt.Sender] = Ratelimit{lastcall: time.Now(), warningmsg: id.EventID("0")}
			InvokeCommand(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, evt.RoomID, evt) //success
		}

	// Nice feature, the user can figure out what we are if he pings us with random text.
	} else if evt.Content.AsMessage().Format == "org.matrix.custom.html" {
		if strings.HasPrefix(evt.Content.AsMessage().FormattedBody, "<a href=\"https://matrix.to/#/"+cmdhdlr.client.UserID.String()) {
			cmdhdlr.BotPrint(BotPrintSimple(evt.RoomID, "You need help? Type in >"+cmdhdlr.prefix+"help<."))
		}
	}

	// Do not put this in an else clause! Bridge messages wont get to this place otherwise!
	cmdhdlr.funcregister.InvokeFunctions(CommandArgs{cmdhdlr, evt.RoomID, evt.Sender, argc, argv, evt.RoomID, evt, commandsInternal{}})
}
