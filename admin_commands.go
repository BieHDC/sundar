package main

import (
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strconv"
	"time"
	"strings"
)

func botJoinRoom(cmdhdlr *CommandHandler, targetroom id.RoomID) bool {
	_, err := cmdhdlr.client.JoinRoomByID(targetroom)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "Join Event Error 1 for room >"+targetroom.String()+"<:"+err.Error())
		return false
	} else {
		cmdhdlr.join_room_event(targetroom)
		BotNotifyEventsChannel(cmdhdlr, "Join Event Success:"+targetroom.String())
		return true
	}
}

func HandleBotForceJoinRoom(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		targetroom, err := parseRoomID(argv[1])
		if err == nil && targetroom != nil {
			BotReplyMsg(cmdhdlr, statusroom, "Joining Room >"+targetroom.String()+"<.")
			botJoinRoom(cmdhdlr, targetroom.RoomID())
		} else {
			BotReplyMsg(cmdhdlr, statusroom, "Invalid RoomID.")
			return false
		}
	}
	return true
}

func botLeaveRoom(cmdhdlr *CommandHandler, targetroom id.RoomID) {
	// You first need to leave it before you can forget it
	_, err := cmdhdlr.client.LeaveRoom(targetroom)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "botLeaveRoom Error 1:"+err.Error())
	}
	time.Sleep(1 * time.Second)
	_, err = cmdhdlr.client.ForgetRoom(targetroom)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "botLeaveRoom Error 2:"+err.Error())
	}
	time.Sleep(1 * time.Second)
}

func HandleBotForceLeaveRoom(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		var targetroom_str string
		if argv[1] == "this" {
			targetroom_str = room.String()
		} else {
			targetroom_str = argv[1]
		}

		targetroom, err := parseRoomID(targetroom_str)
		if err == nil && targetroom != nil {
			if cmdhdlr.userHasPowerlevel(targetroom.RoomID(), sender) >= CommandAdmin {
				BotNotifyEventsChannel(cmdhdlr, "Leaving room "+targetroom.String())
				botLeaveRoom(cmdhdlr, targetroom.RoomID())
			} else {
				BotReplyMsg(cmdhdlr, statusroom, "You need to be an Administrator in >" + targetroom.String() + "< to kick the bot.")
			}
		} else {
			BotNotifyEventsChannel(cmdhdlr, "Error leaving room "+err.Error())
			BotReplyMsg(cmdhdlr, statusroom, "Invalid RoomID.")
			return false
		}
	}
	return true
}

func HandleBotLeaveEmptyRooms(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	left_rooms_count := 0
	joinedrooms, err := cmdhdlr.client.JoinedRooms()
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "HandleBotLeaveEmptyRooms Error 1:"+err.Error())
		return false
	} else {
		for _, targetroom := range joinedrooms.JoinedRooms {
			joinedusers, err := cmdhdlr.client.JoinedMembers(targetroom)
			if err != nil {
				BotNotifyEventsChannel(cmdhdlr, "HandleBotLeaveEmptyRooms Error 2:"+err.Error())
			} else {
				if len(joinedusers.Joined) <= 1 {
					botLeaveRoom(cmdhdlr, targetroom)
					left_rooms_count++
				}
			}
		}
	}
	BotReplyMsg(cmdhdlr, statusroom, "Left "+strconv.Itoa(left_rooms_count)+" Rooms")
	return true
}

func HandleBotBroadcastMessage(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		joinedrooms, err := cmdhdlr.client.JoinedRooms()
		if err != nil {
			BotNotifyEventsChannel(cmdhdlr, "HandleBotBroadcastMessage Error 1:"+err.Error())
		} else {
			for _, targetroom := range joinedrooms.JoinedRooms {
				_, err := cmdhdlr.client.SendNotice(targetroom, argv[1])
				if err != nil {
					BotNotifyEventsChannel(cmdhdlr, "HandleBotBroadcastMessage Error 2:"+err.Error())
				}
			}
		}
	}
	return true
}

// Goodie for Clusters
func HandleBotSayMessage(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		BotReplyMsgPinged(cmdhdlr, sender, room, "", " says: " + strings.Join(argv[1:], " "))
	}
	return true
}

func HandleSetDisplayName(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		err := cmdhdlr.client.SetDisplayName(argv[1])
		if err != nil {
			BotNotifyEventsChannel(cmdhdlr, "HandleSetDisplayName Error:"+err.Error())
		}
	}
	return true
}

func HandleSetAvatar(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc == 2 {
		// Clear Profile Picture
		if argv[1] == "clear" {
			err := cmdhdlr.client.SetAvatarURL(id.ContentURI{})
			if err != nil {
				BotNotifyEventsChannel(cmdhdlr, "HandleSetAvatar Error 1:"+err.Error())
				return false
			}
		} else {
		// Set Profile Picture to URI
			contenturi, err := id.ParseContentURI(argv[1])
			if err != nil {
				BotNotifyEventsChannel(cmdhdlr, "HandleSetAvatar Error 2:"+err.Error())
				return false
			} else {
				err = cmdhdlr.client.SetAvatarURL(contenturi)
				if err != nil {
					BotNotifyEventsChannel(cmdhdlr, "HandleSetAvatar Error 3:"+err.Error())
					return false
				}
			}
		}
	} else {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	}
	return true
}

func HandleSetBotRoomAppearance(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 3 {
		cmdhdlr.internelPrintUsage(argv[0], statusroom)
		return false
	} else {
		var targetroom_str string
		if argv[1] == "this" {
			targetroom_str = room.String()
		} else {
			targetroom_str = argv[1]
		}

		targetroom, err := parseRoomID(targetroom_str)
		if err == nil && targetroom != nil {
			if cmdhdlr.userHasPowerlevel(targetroom.RoomID(), sender) >= CommandAdmin {
				var profilepicture id.ContentURIString
				// Clear Profile Picture
				if argv[2] == "clear" {
					profilepicture = ""
				// New Profile Picture
				} else {
					contenturi, err := id.ParseContentURI(argv[2])
					if err != nil {
						BotReplyMsg(cmdhdlr, statusroom, "Invalid Content URI.")
						return false
					}
					profilepicture = contenturi.CUString()
				}

				newname := "" //if its left empty its going to reset
				if argc < 4 { 
					newname = strings.Join(argv[3:], " ") //otherwise we join the rest together as name
				}

				membershipevent := event.MemberEventContent{
					Membership:  event.MembershipJoin,
					Displayname: newname,
					AvatarURL:   profilepicture,
				}
				_, err := cmdhdlr.client.SendStateEvent(room, event.StateMember, cmdhdlr.client.UserID.String(), membershipevent)
				if err != nil {
					BotReplyMsg(cmdhdlr, statusroom, "HandleSetBotRoomAppearance Error:"+err.Error())
				} else {
					BotReplyMsg(cmdhdlr, statusroom, "New Bot Room appearance has been set.")
				}
			} else {
				BotReplyMsg(cmdhdlr, statusroom, "You need to be an Administrator in >" + targetroom.String() + "< to change the bot appearance.")
			}
		} else {
			BotNotifyEventsChannel(cmdhdlr, "Error leaving room "+err.Error())
			BotReplyMsg(cmdhdlr, statusroom, "Invalid RoomID.")
			return false
		}
	}
	return true
}
