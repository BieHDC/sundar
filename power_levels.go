package main

import (
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strings"
	"strconv"
	"fmt"
)

// this is the room power levels changed callback
func (cmdhdlr *CommandHandler) PowerLevelsChange(targetroom id.RoomID, newpowerlevels map[id.UserID]int) {
	cmdhdlr.masters[targetroom] = newpowerlevels
}

func (cmdhdlr *CommandHandler) loadRoomsAndPowerlevelsFromUserdata() {
	if cmdhdlr.client == nil {
		fmt.Println("!!NO VALID CLIENT FOUND, ASSUMING TEST MODE!!")
		return
	}

	resp, err := cmdhdlr.client.JoinedRooms()
	if err != nil {
		botNotifyEventsChannel(cmdhdlr, "Failed to get joined rooms!")
		return
	}

	cmdhdlr.joined_rooms = resp.JoinedRooms
	for room := range cmdhdlr.joined_rooms {
		ret := event.PowerLevelsEventContent{}
		err := cmdhdlr.client.StateEvent(cmdhdlr.joined_rooms[room], event.StatePowerLevels, "", &ret)
		if err != nil {
			botNotifyEventsChannel(cmdhdlr, "Failed to get power levels: "+err.Error())
			continue
		}

		cmdhdlr.masters[cmdhdlr.joined_rooms[room]] = ret.Users
	}
}

func (cmdhdlr *CommandHandler) loadPowerLevelsOverrides() {
	cmdhdlr.accountstore.FetchData(cmdhdlr.client, "powerlevelsoverrides_1", &cmdhdlr.alloverrides)
}

func (cmdhdlr *CommandHandler) savePowerLevelsOverrides() {
	cmdhdlr.accountstore.StoreData(cmdhdlr.client, "powerlevelsoverrides_1", &cmdhdlr.alloverrides)
}

func doPowerlevelOverride(cmdhdlr *CommandHandler, sender id.UserID, targetroom id.RoomID, commandname string, newpowerlevel string, statusroom id.RoomID) BotReply {
	powerlevel, err := strconv.Atoi(strings.TrimSpace(newpowerlevel))
	if err != nil {
		return BotPrintSimple(statusroom, "Could not convert >" + newpowerlevel + "< to an integer.")
	}

	command, ok := cmdhdlr.allcommands[commandname]
	if !ok {
		return BotPrintSimple(statusroom, "Command >"+commandname+"< not found.")
	}

	if !cmdhdlr.hasUserRequiredPowerlevel(targetroom, sender, commandname, command.RequiredPowerlevel) {
		return BotPrintSimple(statusroom, "You do not have permission to change the powerlevel of a command(" + 
			strconv.Itoa(cmdhdlr.needsAtLeastPowerlevel(targetroom, commandname, command.RequiredPowerlevel)) + 
			") with higher level than you(" + strconv.Itoa(cmdhdlr.userHasPowerlevel(targetroom, sender)) + ").")
		
	}

	// we have the power, the command and the powerlevel, lets go
	//3. change the internal map powerlevel
	cmdhdlr.alloverrides[commandname][targetroom] = powerlevel
	//4. publish the change to the userdata
	cmdhdlr.savePowerLevelsOverrides()
	//5. invalidate help text cache
	cmdhdlr.helptextcache.invalidateHelpTextCacheForRoom(targetroom)

	return BotPrintSimple(statusroom, "Set new powerlevel of >" + newpowerlevel + "< for command >" + commandname + ".")
}

func HandleCommandPowerlevelOverride(ca CommandArgs) BotReply {
	if ca.argc == 3 {
		return doPowerlevelOverride(ca.cmdhdlr, ca.sender, ca.room, ca.argv[1], ca.argv[2], ca.statusroom)
	} else if ca.argc == 4 {
		targetroom, err := parseRoomID(ca.argv[1])
		if err == nil && targetroom != nil {
			return doPowerlevelOverride(ca.cmdhdlr, ca.sender, targetroom.RoomID(), ca.argv[2], ca.argv[3], ca.statusroom)
		} else {
			return BotPrintSimple(ca.statusroom, "Error parsing RoomID: " + err.Error() + ".")
		}
	} else {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}
}

func HandleCommandPowerlevelBlock(ca CommandArgs) BotReply {
	if ca.argc == 2 {
		return doPowerlevelOverride(ca.cmdhdlr, ca.sender, ca.room, ca.argv[1], "100", ca.statusroom)
	} else if ca.argc == 3 {
		targetroom, err := parseRoomID(ca.argv[1])
		if err == nil && targetroom != nil {
			return doPowerlevelOverride(ca.cmdhdlr, ca.sender, targetroom.RoomID(), ca.argv[2], "100", ca.statusroom)
		} else {
			return BotPrintSimple(ca.statusroom, "Error parsing RoomID: " + err.Error() + ".")
		}
	} else {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}
}

func HandleCommandPowerlevelOverrideReset(ca CommandArgs) BotReply {
	var targetroom id.RoomID
	var targetcommand string

	if ca.argc == 2 {
		targetroom = ca.room
		targetcommand = ca.argv[1]
	} else if ca.argc == 3 {
		roomid, err := parseRoomID(ca.argv[1])
		if err != nil || roomid == nil {
			return BotPrintSimple(ca.statusroom, "Error parsing RoomID: " + err.Error() + ".")
		}
		targetroom = roomid.RoomID()
		targetcommand = ca.argv[2]
	} else {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	command, ok := ca.cmdhdlr.allcommands[targetcommand]
	if !ok {
		return BotPrintSimple(ca.statusroom, "Command >"+targetcommand+"< not found.")
	}

	if !ca.cmdhdlr.hasUserRequiredPowerlevel(targetroom, ca.sender, targetcommand, command.RequiredPowerlevel) {
		return BotPrintSimple(ca.statusroom, "You do not have permission to change the powerlevel of a command(" + 
			strconv.Itoa(ca.cmdhdlr.needsAtLeastPowerlevel(targetroom, targetcommand, command.RequiredPowerlevel)) + 
			") with higher level than you(" + strconv.Itoa(ca.cmdhdlr.userHasPowerlevel(targetroom, ca.sender)) + ").")
	}

	//3. change the internal map powerlevel
	delete(ca.cmdhdlr.alloverrides[targetcommand], targetroom)
	//4. publish the change to the userdata
	ca.cmdhdlr.savePowerLevelsOverrides()
	//5. invalidate help text cache
	ca.cmdhdlr.helptextcache.invalidateHelpTextCacheForRoom(targetroom)
	return BotPrintSimple(ca.statusroom, "Powerlevel for command >" + targetcommand + "< has been reset to its default >" + strconv.Itoa(command.RequiredPowerlevel) + "<.")
}

func HandleCommandPowerlevelOverrideResetForAllCommandsInRoom(ca CommandArgs) BotReply {
	targetroom := ca.room
	if ca.argc == 2 {
		roomid, err := parseRoomID(ca.argv[1])
		if err != nil || roomid == nil {
			return BotPrintSimple(ca.statusroom, "Error parsing RoomID: " + err.Error() + ".")
		}
		targetroom = roomid.RoomID()
	}

	failedcommands := ""
	for commandname, cmdint := range ca.cmdhdlr.allcommands { 
		if ca.cmdhdlr.hasUserRequiredPowerlevel(targetroom, ca.sender, commandname, cmdint.RequiredPowerlevel) {
			//3. change the internal map powerlevel
			delete(ca.cmdhdlr.alloverrides[commandname], targetroom)
		} else {
			_, exists := ca.cmdhdlr.alloverrides[commandname][targetroom] // If we dont do that it gets spammy real quick
			if exists {
				failedcommands += commandname + ", "
			}
		}
	}
	//4. publish the change to the userdata
	ca.cmdhdlr.savePowerLevelsOverrides()
	//5. invalidate help text cache
	ca.cmdhdlr.helptextcache.invalidateHelpTextCacheForRoom(targetroom)

	if failedcommands == "" {
		return BotPrintSimple(ca.statusroom, "All command powerlevel overrides in the room have been reset to their defaults.")
	} else {
		failedcommands = strings.TrimSuffix(failedcommands, ", ") // Cope and Seethe Reader
		return BotPrintSimple(ca.statusroom, "The following command powerlevel overrides in the room could not have been reset due to lack of permission: " + failedcommands)
	}
}


func (cmdhdlr *CommandHandler) needsAtLeastPowerlevel(targetroom id.RoomID, commandname string, requiredpowerlevel int) int {
	overridepowerlevel, overrideexists := cmdhdlr.alloverrides[commandname][targetroom]
	// When an override exists, the default is ignored
	
	// If a local override exists
	if overrideexists {
		return overridepowerlevel
	// Otherwise we check against the default
	} else {
		return requiredpowerlevel
	}
}

func (cmdhdlr *CommandHandler) hasUserRequiredPowerlevel(targetroom id.RoomID, sender id.UserID, commandname string, requiredpowerlevel int) bool {
	userpower, haspower := cmdhdlr.masters[targetroom][sender]
	neededpowerlevel := cmdhdlr.needsAtLeastPowerlevel(targetroom, commandname, requiredpowerlevel)
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
	// You are too weak
	return false
}

func (cmdhdlr *CommandHandler) userHasPowerlevel(targetroom id.RoomID, sender id.UserID) int {
	userpower, haspower := cmdhdlr.masters[targetroom][sender]
	if !haspower {
		return 0
	}

	if sender != cmdhdlr.botmaster {
		// Keep people from impersonating the botmaster
		if userpower > 100 {
			userpower = 100
		}
	}
	
	return userpower
}
