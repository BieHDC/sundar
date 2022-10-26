package main

import (
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"time"
	"encoding/json"
	"errors"
)


func botJoinRoom(cmdhdlr *CommandHandler, targetroom id.RoomID) (bool, error) {
	_, err := cmdhdlr.client.JoinRoomByID(targetroom)
	if err != nil {
		return false, errors.New("Join Event Error 1 for room >"+targetroom.String()+"<:"+err.Error())
	}

	err = cmdhdlr.join_room_event(targetroom)
	if err != nil {
		return false, errors.New("Join Event Error 2 for room >"+targetroom.String()+"<:"+err.Error())
	} else {
		return true, errors.New("Join Event Success:"+targetroom.String())
	}
}

func botLeaveRoom(cmdhdlr *CommandHandler, targetroom id.RoomID) (bool, error) {
	// You first need to leave it before you can forget it
	// Might be possible we somehow already left but did not yet forget
	//  so we always do both
	var errormessage string

	_, err := cmdhdlr.client.LeaveRoom(targetroom)
	if err != nil {
		errormessage = "botLeaveRoom Error 1:"+err.Error()
	}
	time.Sleep(1 * time.Second)

	_, err = cmdhdlr.client.ForgetRoom(targetroom)
	if err != nil {
		if errormessage != "" {
			errormessage += " -- botLeaveRoom Error 2:"+err.Error()
		} else {
			errormessage = "botLeaveRoom Error 2:"+err.Error()
		}
	}
	time.Sleep(1 * time.Second)

	if errormessage != "" {
		return false, errors.New(errormessage)
	} else {
		return true, nil
	}
}


func (cmdhdlr *CommandHandler) join_room_event(targetroom id.RoomID) error {
	//add the new room to our list
	cmdhdlr.joined_rooms = append(cmdhdlr.joined_rooms, targetroom)
	ret := event.PowerLevelsEventContent{}
	err := cmdhdlr.client.StateEvent(targetroom, event.StatePowerLevels, "", &ret)
	if err != nil {
		return errors.New("Failed to get power levels: "+err.Error())
	}
	cmdhdlr.masters[targetroom] = ret.Users
	return nil
}

func (cmdhdlr *CommandHandler) leave_room_event(targetroom id.RoomID) {
	//wipe any userdata connnected to the room
	emptyjson := []byte(`{}`)
	var emptyjsondat map[string]interface{}
    json.Unmarshal(emptyjson, &emptyjsondat)
	for commandname, _ := range cmdhdlr.allcommands {
		cmdhdlr.accountstore.StoreData(cmdhdlr.client, "cmdplroomoverridefor:"+commandname, &emptyjsondat)
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




