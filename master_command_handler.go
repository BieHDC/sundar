//todo:
// before v2

// sometime in the future:
//	bring back the cursed starboard? yes
//	ratelimiter to delete spammed messages? maybe out of scope since a moderation bot should do that.
//	scan for :oof: patterns and post the emoji (needs custom emoji pics support from matrix)
//	add option for the botmaster to set status (needs matrix side support, name and pfp global and per room done)
//  tie up e2ee support once proper support for bot-e2ee is implemented


package main

import (
	//"fmt"
	"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strconv"
	"strings"
	"encoding/json"
	"time"
)

// You can also provide a number directly btw
const CommandAnyone 	int = 0
const CommandModerator 	int = 50
const CommandAdmin 		int = 100
const CommandMaster 	int = 1000

//We still pass it in for a small specific amount of calls, but have it friendlier to use with if we extrace the sender and room
type CommandCallback func(*CommandHandler, id.RoomID, id.UserID, int, []string, id.RoomID, *event.Event) bool
type commandsInternal struct {
	Type_              			string          	// What kind of command it is. For autohelp.
	Description        			string          	// What it does. For autohelp.
	Usage 						string 				// For the "usage" command to be picked up.
	RequiredPowerlevel 			int             	// The minimum powerlevel a user needs for the command.
	Targetfunc         			CommandCallback 	// The function to call if we match.
}

type CommandHandler struct {
	client         *mautrix.Client
	eventsChannel  id.RoomID
	eventsToStdout bool
	prefix         string

	joined_rooms []id.RoomID
	accountstore *AccountStorage

	funcregister 	*FunctionRegister
	clusterregister *ClusterRegister
	echoregister	*EchoRegister

	allcommands 	map[string]commandsInternal
	alloverrides 	map[string]map[id.RoomID]int // command -> room/powerlevel // If the command has a different powerlevel in a room

	botmaster  	id.UserID // Untouchable Admin, Botmaster
	masters 	map[id.RoomID]map[id.UserID]int

	serverstatus 	*ServerStatus
}

func NewCommandHandler(client *mautrix.Client, prefix string, master id.UserID, eventsChannel id.RoomID, eventsToStdout bool, accountstoreprefix string, accountstoredebug bool) *CommandHandler {
	cmdhdlr := CommandHandler{
		client:         client,
		eventsChannel:  eventsChannel,
		eventsToStdout: eventsToStdout,
		prefix:         prefix,

		allcommands: 	make(map[string]commandsInternal),
		alloverrides: 	make(map[string]map[id.RoomID]int),
		accountstore: 	NewDefaultAccountStorage(accountstoreprefix, accountstoredebug),

		funcregister: 		NewDefaultFunctionRegister(),
		clusterregister: 	NewClusterHandler(),
		echoregister:		NewEchoRegister(),

		botmaster: 	master,
		masters: 	make(map[id.RoomID]map[id.UserID]int),

		serverstatus: 	NewServerStatus(),
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
	cmdhdlr.AddCommand("broadcast", 		"Broadcast Message to all joined rooms", 	"<Your Message>", 	"Tools", CommandMaster, HandleBotBroadcastMessage)

	// Tweaking
	cmdhdlr.AddCommand("setbotroomappearance", 	"Change the bots displayname and avatar in this room", 	"<this>OR<RoomID> <mxc uri>OR<clear> <New Display Name>", 	"Appearance", CommandAdmin, HandleSetBotRoomAppearance)
	cmdhdlr.AddCommand("setdisplayname", 		"Change the bots displayname", 							"<New Display Name>", 										"Appearance", CommandMaster, HandleSetDisplayName)
	cmdhdlr.AddCommand("setavatar", 			"Change the bots avatar", 								"<mxc uri>OR<clear>", 										"Appearance", CommandMaster, HandleSetAvatar)

	// Power Level Management
	cmdhdlr.AddCommand("poweroverride", 		"Override the required powerlevel for a command for this room", "[roomid] <command> <new power level>", "Powermanagement", CommandAdmin, HandleCommandPowerlevelOverride)
	cmdhdlr.AddCommand("poweroverridereset", 	"Reset the override powerlevel for a command for this room", 	"[roomid] <command>", 					"Powermanagement", CommandAdmin, HandleCommandPowerlevelOverrideReset)
	cmdhdlr.AddCommand("poweroverrideresetall", "Reset all overrides for a room", 								"[roomid]",								"Powermanagement", CommandAdmin, HandleCommandPowerlevelOverrideResetForAllCommandsInRoom)
	cmdhdlr.AddCommand("blockcommand", 			"Quicker way to block a command for all but admins", 			"[roomid] <command>", 					"Powermanagement", CommandAdmin, HandleCommandPowerlevelBlock)

	// Clustermanagement
	cmdhdlr.AddCommand("cluster",	"Manage a Cluster of Rooms",	"<clustercommand> []\nlist/listroom/join/leave/remove/cmd",	"Cluster", CommandAdmin, HandleCluster)
	cmdhdlr.AddCommand("say", 		"Say a message in a room", 		"<Your Message>", 											"Cluster", CommandAdmin, HandleBotSayMessage)

	// Functionmanagement
	cmdhdlr.AddCommand("funcstate",		"Set the State of a Function in a Room", 	"<functionid> <enable>OR<disable> <this>OR<RoomID>", 	"Functionmanagement", CommandAdmin, 	HandleFunctionStatusInRoom)
	cmdhdlr.AddCommand("funcwiperooms",	"Deletes the Roomlist for a Function", 		"<functionid>", 										"Functionmanagement", CommandMaster, HandleFunctionWipeRoomlist)
	cmdhdlr.AddCommand("funclist",		"List all registered Functions", 			"", 													"Functionmanagement", CommandAdmin, 	HandleListKnownFunctions)

	// Echostuff
	cmdhdlr.AddCommand("echoadd", "Add a command that just echos a message", 	"<callsign> <Your Message>", 	"Echomanagement", CommandAdmin, HandleBotFormattedMessageAdd)
	cmdhdlr.AddCommand("echodel", "Remove a command that just echos a message", "<callsign>", 					"Echomanagement", CommandAdmin, HandleBotFormattedMessageRemove)

	// Help
	cmdhdlr.AddCommand("help", 	"Displays this help message", 		"", "Help", CommandAnyone, printHelp)
	cmdhdlr.AddCommand("usage",	"Displays the usage of a command", 	"[command]\n<arg> is a required argument\n[arg] is an optional argument\nThe OR keyword tells you to choose between 2 or more options on a field.", "Help", CommandAnyone, printUsage)
	// any commands need to be added before this line


	// Load Echo Commands
	loadEchos(cmdhdlr, cmdhdlr.echoregister)

	//load power level overrides for commands
	cmdhdlr.loadPowerLevelsOverrides()

	//load cluster data
	loadClusters(cmdhdlr, cmdhdlr.clusterregister)

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
		//BotNotifyEventsChannel(cmdhdlr, "The command with the callsign >" + targetcommand + "< exists already!!")
		return false
	}
}

// Only meant for echodel
func (cmdhdlr *CommandHandler) removeCommand(targetcommand string) {
	delete(cmdhdlr.allcommands, targetcommand)
	cmdhdlr.invalidateHelpTextCacheForAll()
}


func InvokeCommand(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	//fmt.Println(argv)
	command, ok := cmdhdlr.allcommands[argv[0]]
	if ok {
		if cmdhdlr.hasUserRequiredPowerlevel(room, sender, argv[0], &command) {
			return cmdhdlr.allcommands[argv[0]].Targetfunc(cmdhdlr, room, sender, argc, argv, statusroom, evt)
			// The return-bool is needed for cluster commands, if you end up printing helptext, the args are probably wrong and you want to stop right there
		} else {
			complaint := "HandleDispatch: User >" + sender.String() + 
							"< tried to do the admin command >" + argv[0] + 
							"< which requires at least power level " + 
							strconv.Itoa(cmdhdlr.needsAtLeastPowerlevel(room, argv[0], &command))
			BotNotifyEventsChannel(cmdhdlr, complaint)
			BotReplyMsgPinged(cmdhdlr, sender, statusroom, "",
				": You do not have permission to run this command, this incident will be reported")
			return true // User might have the perms to run it in the other rooms in the clusterlist
		}
	} else {
		BotReplyMsg(cmdhdlr, statusroom, "I do not know about this command. Did you check >"+cmdhdlr.prefix+"help<?")
		return false // If the command does not exist however, we should break the loop
	}
	return true
}

type Ratelimit struct {
	lastcall time.Time
	warningmsg id.EventID
}
var ratelimit = make(map[id.UserID]Ratelimit)

const granulatymsg = time.Millisecond*100
const timelimit = time.Second * 3
func (cmdhdlr *CommandHandler) HandleDispatch(evt *event.Event) {
	argv, argc := sanitiseParseCommandWithArguments(evt.Content.AsMessage().Body, cmdhdlr.prefix)
	if strings.HasPrefix(evt.Content.AsMessage().Body, cmdhdlr.prefix) {
		// Simple Ratelimiter
		rl, exists := ratelimit[evt.Sender]
		stamp := rl.lastcall.Add(timelimit).Truncate(granulatymsg).After(time.Now())
		if exists && stamp {
			if rl.warningmsg == id.EventID(0) {
				warnreply := "Cool it! Wait " + timelimit.String() + "."
				warningmsg := BotReplyMsgRet(cmdhdlr, evt.RoomID, warnreply)
				ratelimit[evt.Sender] = Ratelimit{lastcall: time.Now(), warningmsg: warningmsg}
			} else {
				timeleft := (((timelimit)-time.Now().Sub(rl.lastcall))/2).Truncate(granulatymsg)
				newtime := time.Now().Add(timeleft) //we do a little trolling
				ratelimit[evt.Sender] = Ratelimit{lastcall: newtime, warningmsg: rl.warningmsg}
				warnreply := "Cool it more! Wait for " + timeleft.String() + "."
				BotReplyMsgEdited(cmdhdlr, evt.RoomID, warnreply, rl.warningmsg)
				//fixme here we could add the function for the bot to delete the spammed message
			}
			return //we do not serve you
		} else {
			ratelimit[evt.Sender] = Ratelimit{lastcall: time.Now(), warningmsg: id.EventID(0)}
		}

		if argc > 0 {
			InvokeCommand(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, evt.RoomID, evt)
		} else {
			BotReplyMsg(cmdhdlr, evt.RoomID, "There is no such thing as an empty command. That would be ridicilous!")
		}

	// Nice feature, the user can figure out what we are if he pings us with random text.
	} else if evt.Content.AsMessage().Format == "org.matrix.custom.html" {
		if strings.HasPrefix(evt.Content.AsMessage().FormattedBody, "<a href=\"https://matrix.to/#/"+cmdhdlr.client.UserID.String()) {
			BotReplyMsg(cmdhdlr, evt.RoomID, "You need help? Type in >"+cmdhdlr.prefix+"help<.")
		}

	}

	// Do not put this in an else clause! Bridge messages wont get to this place otherwise!
	cmdhdlr.funcregister.InvokeFunctions(cmdhdlr, evt.RoomID, evt.Sender, argc, argv, evt)
}


func (cmdhdlr *CommandHandler) join_room_event(targetroom id.RoomID) {
	//add the new room to our list
	cmdhdlr.joined_rooms = append(cmdhdlr.joined_rooms, targetroom)
	ret := event.PowerLevelsEventContent{}
	err := cmdhdlr.client.StateEvent(targetroom, event.StatePowerLevels, "", &ret)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "Failed to get power levels: "+err.Error())
	} else {
		cmdhdlr.masters[targetroom] = ret.Users
	}
}

func (cmdhdlr *CommandHandler) leave_room_event(targetroom id.RoomID) {
	//wipe any userdata connnected to the room
	emptyjson := []byte(`{}`)
	var emptyjsondat map[string]interface{}
    json.Unmarshal(emptyjson, &emptyjsondat)
	for commandname, _ := range cmdhdlr.allcommands {
		cmdhdlr.StoreData("cmdplroomoverridefor:"+commandname, &emptyjsondat)
	}

	//remove room from our list
	var newrooms []id.RoomID
	for room := range cmdhdlr.joined_rooms {
		if cmdhdlr.joined_rooms[room] != targetroom {
			newrooms = append(newrooms, cmdhdlr.joined_rooms[room])
		}
	}
	cmdhdlr.joined_rooms = newrooms
}
