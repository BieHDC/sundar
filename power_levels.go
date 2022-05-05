package main

import (
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strings"
	"strconv"
)

// this is the room power levels changed callback
func (cmdhdlr *CommandHandler) PowerLevelsChange(targetroom id.RoomID, newpowerlevels map[id.UserID]int) {
	cmdhdlr.masters[targetroom] = newpowerlevels
}

func (cmdhdlr *CommandHandler) loadRoomsAndPowerlevelsFromUserdata() {
	resp, err := cmdhdlr.client.JoinedRooms()
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "Failed to get joined rooms!")
	} else {
		cmdhdlr.joined_rooms = resp.JoinedRooms
		for room := range cmdhdlr.joined_rooms {
			ret := event.PowerLevelsEventContent{}
			err := cmdhdlr.client.StateEvent(cmdhdlr.joined_rooms[room], event.StatePowerLevels, "", &ret)
			if err != nil {
				BotNotifyEventsChannel(cmdhdlr, "Failed to get power levels: "+err.Error())
			} else {
				cmdhdlr.masters[cmdhdlr.joined_rooms[room]] = ret.Users
			}
		}
	}
}

func (cmdhdlr *CommandHandler) loadPowerLevelsOverrides() {
	cmdhdlr.FetchData("powerlevelsoverrides_1", &cmdhdlr.alloverrides)
}

func (cmdhdlr *CommandHandler) savePowerLevelsOverrides() {
	cmdhdlr.StoreData("powerlevelsoverrides_1", &cmdhdlr.alloverrides)
}

func doPowerlevelOverride(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, targetroom id.RoomID, commandname string, newpowerlevel string, statusroom id.RoomID) bool {
	command, ok := cmdhdlr.allcommands[commandname]
	if ok {
		if cmdhdlr.hasUserRequiredPowerlevel(targetroom, sender, commandname, &command) {
			powerlevel, err := strconv.Atoi(strings.TrimSpace(newpowerlevel))
			if err == nil {
				// we have the power, the command and the powerlevel, lets go
				//3. change the internal map powerlevel
				cmdhdlr.alloverrides[commandname][targetroom] = powerlevel
				//4. publish the change to the userdata
				cmdhdlr.savePowerLevelsOverrides()
				//5. invalidate help text cache
				cmdhdlr.invalidateHelpTextCacheForRoom(targetroom)
				BotReplyMsg(cmdhdlr, statusroom, "Set new powerlevel of >" + newpowerlevel + "< for command >" + commandname + ".")
			} else {
				BotReplyMsg(cmdhdlr, statusroom, "Could not convert >" + newpowerlevel + "< to an integer.")
				return false
			}
		} else {
			BotReplyMsg(cmdhdlr, statusroom, "You do not have permission to change the powerlevel of a command(" + 
				strconv.Itoa(cmdhdlr.needsAtLeastPowerlevel(targetroom, commandname, &command)) + 
				") with higher level than you(" + strconv.Itoa(cmdhdlr.userHasPowerlevel(targetroom, sender)) + ").")
		}
	} else {
		BotReplyMsg(cmdhdlr, statusroom, "Command >"+commandname+"< not found.")
		return false
	}
	return true
}

func HandleCommandPowerlevelOverride(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc == 3 {
		return doPowerlevelOverride(cmdhdlr, room, sender, room, argv[1], argv[2], statusroom)
	} else if argc == 4 {
		targetroom, err := parseRoomID(argv[1])
		if err == nil && targetroom != nil {
			return doPowerlevelOverride(cmdhdlr, room, sender, targetroom.RoomID(), argv[2], argv[3], statusroom)
		} else {
			BotReplyMsg(cmdhdlr, statusroom, "Error parsing RoomID: " + err.Error() + ".")
			return false
		}
	} else {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	}
	return true
}

func HandleCommandPowerlevelBlock(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc == 2 {
		return doPowerlevelOverride(cmdhdlr, room, sender, room, argv[1], "100", statusroom)
	} else if argc == 3 {
		targetroom, err := parseRoomID(argv[1])
		if err == nil && targetroom != nil {
			return doPowerlevelOverride(cmdhdlr, room, sender, targetroom.RoomID(), argv[2], "100", statusroom)
		} else {
			BotReplyMsg(cmdhdlr, statusroom, "Error parsing RoomID: " + err.Error() + ".")
			return false
		}
	} else {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	}
	return true
}

func HandleCommandPowerlevelOverrideReset(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	var targetroom id.RoomID
	var targetcommand string

	if argc == 2 {
		targetroom = room
		targetcommand = argv[1]
	} else if argc == 3 {
		roomid, err := parseRoomID(argv[1])
		if err != nil || roomid == nil {
			BotReplyMsg(cmdhdlr, statusroom, "Error parsing RoomID: " + err.Error() + ".")
			return false
		}
		targetroom = roomid.RoomID()
		targetcommand = argv[2]
	} else {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	}

	command, ok := cmdhdlr.allcommands[targetcommand]
	if ok {
		if cmdhdlr.hasUserRequiredPowerlevel(targetroom, sender, targetcommand, &command) {
			//3. change the internal map powerlevel
			delete(cmdhdlr.alloverrides[targetcommand], targetroom)
			//4. publish the change to the userdata
			cmdhdlr.savePowerLevelsOverrides()
			//5. invalidate help text cache
			cmdhdlr.invalidateHelpTextCacheForRoom(targetroom)
			BotReplyMsg(cmdhdlr, statusroom, "Powerlevel for command >" + targetcommand + "< has been reset to its default >" + strconv.Itoa(command.RequiredPowerlevel) + "<.")
		} else {
			BotReplyMsg(cmdhdlr, statusroom, "You do not have permission to change the powerlevel of a command(" + 
				strconv.Itoa(cmdhdlr.needsAtLeastPowerlevel(targetroom, targetcommand, &command)) + 
				") with higher level than you(" + strconv.Itoa(cmdhdlr.userHasPowerlevel(targetroom, sender)) + ").")
		}
	} else {
		BotReplyMsg(cmdhdlr, statusroom, "Command >"+targetcommand+"< not found.")
		return false
	}
	return true
}

func HandleCommandPowerlevelOverrideResetForAllCommandsInRoom(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	targetroom := room
	if argc == 2 {
		roomid, err := parseRoomID(argv[1])
		if err != nil || roomid == nil {
			BotReplyMsg(cmdhdlr, statusroom, "Error parsing RoomID: " + err.Error() + ".")
			return false
		}
		targetroom = roomid.RoomID()
	}

	failedcommands := ""
	for commandname, cmdint := range cmdhdlr.allcommands { 
		if cmdhdlr.hasUserRequiredPowerlevel(targetroom, sender, commandname, &cmdint) {
			//3. change the internal map powerlevel
			delete(cmdhdlr.alloverrides[commandname], targetroom)
		} else {
			_, exists := cmdhdlr.alloverrides[commandname][targetroom] // If we dont do that it gets spammy real quick
			if exists {
				failedcommands += commandname + ", "
			}
		}
	}
	//4. publish the change to the userdata
	cmdhdlr.savePowerLevelsOverrides()
	//5. invalidate help text cache
	cmdhdlr.invalidateHelpTextCacheForRoom(targetroom)

	if failedcommands == "" {
		BotReplyMsg(cmdhdlr, statusroom, "All command powerlevel overrides in the room have been reset to their defaults.")
	} else {
		failedcommands = strings.TrimSuffix(failedcommands, ", ") // Cope and Seethe Reader
		BotReplyMsg(cmdhdlr, statusroom, "The following command powerlevel overrides in the room could not have been reset due to lack of permission: " + failedcommands)
	}
	return true
}


func (cmdhdlr *CommandHandler) needsAtLeastPowerlevel(targetroom id.RoomID, commandname string, command *commandsInternal) int {
	overridepowerlevel, overrideexists := cmdhdlr.alloverrides[commandname][targetroom]
	// When an override exists, the default is ignored
	
	// If a local override exists
	if overrideexists {
		return overridepowerlevel
	// Otherwise we check against the default
	} else {
		return command.RequiredPowerlevel
	}
}

func (cmdhdlr *CommandHandler) hasUserRequiredPowerlevel(targetroom id.RoomID, sender id.UserID, commandname string, command *commandsInternal) bool {
	userpower, haspower := cmdhdlr.masters[targetroom][sender]
	neededpowerlevel := cmdhdlr.needsAtLeastPowerlevel(targetroom, commandname, command)
	// When an override exists, the default is ignored
	
	if !haspower {
		userpower = 0
	}
	// Keep people from impersonating the botmaster
	if userpower > CommandAdmin {
		userpower = CommandAdmin
	}
	if userpower >= neededpowerlevel {
		return true
	}
	// Master Admin
	if sender == cmdhdlr.botmaster {
		return true
	}
	return false
}

func (cmdhdlr *CommandHandler) userHasPowerlevel(targetroom id.RoomID, sender id.UserID) int {
	userpower, haspower := cmdhdlr.masters[targetroom][sender]
	if !haspower {
		return 0
	} else {
		if sender != cmdhdlr.botmaster {
			// Keep people from impersonating the botmaster
			if userpower > 100 {
				userpower = 100
			}
		}
		return userpower
	}
}
