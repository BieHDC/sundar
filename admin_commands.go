package main

import (
	//"maunium.net/go/mautrix"
	//event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strconv"
	"strings"
)

func HandleBotForceJoinRoom(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	targetroom, err := parseRoomID(ca.argv[1])
	if err != nil {
		return BotPrintSimple(ca.statusroom, "Invalid RoomID.")
	}

	success, err := botJoinRoom(ca.cmdhdlr, targetroom.RoomID())
	if success {
		return BotPrintSimple(ca.statusroom, "Joining Room >"+targetroom.String()+"<.")
	} else {
		return BotPrintSimple(ca.statusroom, err.Error())
	}
}

func HandleBotForceLeaveRoom(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	var targetroom_str string
	if ca.argv[1] == "this" {
		targetroom_str = ca.room.String()
	} else {
		targetroom_str = ca.argv[1]
	}
	targetroom, err := parseRoomID(targetroom_str)
	if err != nil {
		return BotPrintSimple(ca.statusroom, "Invalid RoomID.")
	}

	if !(ca.cmdhdlr.userHasPowerlevel(targetroom.RoomID(), ca.sender) >= CommandAdmin) {
		return BotPrintSimple(ca.statusroom, "You need to be an Administrator in >" + targetroom.String() + "< to kick the bot.")
	}

	success, err := botLeaveRoom(ca.cmdhdlr, targetroom.RoomID())
	if success {
		return BotPrintSimple(ca.statusroom, "Leaving room "+targetroom.String())
	} else {
		return BotPrintSimple(ca.statusroom, err.Error())
	}
}

func HandleBotLeaveEmptyRooms(ca CommandArgs) BotReply {
	left_rooms_count := 0
	joinedrooms, err := ca.cmdhdlr.client.JoinedRooms()
	if err != nil {
		return BotPrintSimple(ca.statusroom, "HandleBotLeaveEmptyRooms Error 1:"+err.Error())
	}

	var errorsfromleave []string
	for _, targetroom := range joinedrooms.JoinedRooms {
		joinedusers, err := ca.cmdhdlr.client.JoinedMembers(targetroom)
		if err != nil {
			errorsfromleave = append(errorsfromleave, "HandleBotLeaveEmptyRooms Error 2:"+err.Error())
			continue
		}

		if len(joinedusers.Joined) <= 1 {
			success, err := botLeaveRoom(ca.cmdhdlr, targetroom)
			if !success {
				errorsfromleave = append(errorsfromleave, err.Error())
			} else {
				left_rooms_count++
			}
		}
	}

	errorsformatted := "The folloing errors have occurred while leaving:\n"
	if len(errorsfromleave) > 0 {
		for _, errs := range errorsfromleave {
			errorsformatted += errs + "\n"
		}
		return BotPrintSimple(ca.statusroom, "Left "+strconv.Itoa(left_rooms_count)+" Rooms\n" + errorsformatted)
	}

	return BotPrintSimple(ca.statusroom, "Left "+strconv.Itoa(left_rooms_count)+" Rooms")
}

func HandleBotBroadcastMessage(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	joinedrooms, err := ca.cmdhdlr.client.JoinedRooms()
	if err != nil {
		return BotPrintSimple(ca.statusroom, "HandleBotBroadcastMessage Error 1:"+err.Error())
	}

	chain := BotPrintNothing()
	message := strings.Join(ca.argv[1:], " ")
	for _, targetroom := range joinedrooms.JoinedRooms {
		roommsg := BotPrintSimple(targetroom, message)
		BotPrintAppend(&chain, &roommsg)
	}

	return chain
}

func HandleSetDisplayName(ca CommandArgs) BotReply {
	if ca.argc < 2 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	err := ca.cmdhdlr.client.SetDisplayName(strings.Join(ca.argv[1:], " "))
	if err != nil {
		return BotPrintSimple(ca.statusroom, "HandleSetDisplayName Error:"+err.Error())
	}

	return BotPrintSimple(ca.statusroom, "Successfully set displayname")
}

func HandleSetAvatar(ca CommandArgs) BotReply {
	if ca.argc != 2 {
		return BotPrintSimple(ca.statusroom, ca.self.Usage)
	}

	// Clear Profile Picture
	if ca.argv[1] == "clear" {
		err := ca.cmdhdlr.client.SetAvatarURL(id.ContentURI{})
		if err != nil {
			return BotPrintSimple(ca.statusroom, "HandleSetAvatar Error 1:"+err.Error())
		}
		return BotPrintSimple(ca.statusroom, "Cleared profile picture")
	}

	// Set Profile Picture to URI
	contenturi, err := id.ParseContentURI(ca.argv[1])
	if err != nil {
		return BotPrintSimple(ca.statusroom, "HandleSetAvatar Error 2:"+err.Error())
	}
	err = ca.cmdhdlr.client.SetAvatarURL(contenturi)
	if err != nil {
		return BotPrintSimple(ca.statusroom, "HandleSetAvatar Error 3:"+err.Error())
	}

	return BotPrintSimple(ca.statusroom, "Successfully set profile picture")
}
