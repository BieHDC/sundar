package main

import (
	"fmt"
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
)

type BotReply struct {
	Print			bool
	Room 			id.RoomID
	Msg 			string

	Pinged 			*id.UserID //if _user_ != id.UserID(0) -> pinged -> Msg++Pinged++Msg2
	Msg2 			string

	Replyto 		*event.Event //if _event_ != nil -> set reply

	next			*BotReply //chain prints
}

func BotPrintNothing() BotReply {
	return BotReply{Print: false}
}

func BotPrintSimple(room id.RoomID, msg string) BotReply {
	return BotReply{Print: true, Room: room, Msg: msg}
}

func BotPrintFormatted(room id.RoomID, msg_normal string, msg_formatted string) BotReply {
	return BotReply{Print: true, Room: room, Msg: msg_normal, Msg2: msg_formatted}
}

func BotPrintPinged(room id.RoomID, user *id.UserID, msgpre string, msgpost string) BotReply {
	return BotReply{Print: true, Room: room, Msg: msgpre, Pinged: user, Msg2: msgpost}
}

func BotPrintReplied(room id.RoomID, msg string, replyto *event.Event) BotReply {
	return BotReply{Print: true, Room: room, Msg: msg, Replyto: replyto}
}

func BotPrintAppend(head *BotReply, appendee *BotReply) {
	if head.next != nil {
		BotPrintAppend(head.next, appendee)
	} else {
		head.next = appendee
	}
}

func (cmdhdlr *CommandHandler) BotPrint(br BotReply) {
	defer func() {
		if br.next != nil {
			cmdhdlr.BotPrint(*br.next)
		}
	}()

	if !br.Print {
		return
	}
	//fmt.Printf("%#v\n", br)

	dispatch := event.MessageEventContent{
		MsgType:       event.MsgNotice,
	}

	if br.Replyto != nil {
		dispatch.SetReply(br.Replyto)
	}

	if br.Pinged == nil {
		dispatch.Body = br.Msg
		if br.Msg2 != "" { // Its a formatted message, but not a pinged one
			dispatch.Format = "org.matrix.custom.html"
			dispatch.FormattedBody = br.Msg2
		}
	} else {
		dispatch.Body = br.Msg + br.Pinged.String() + br.Msg2

		pingstring := formatMessageForPing(cmdhdlr, *br.Pinged)
		if pingstring != "" {
			dispatch.Format = "org.matrix.custom.html"
			dispatch.FormattedBody = br.Msg + pingstring + br.Msg2
		}
	}

	cmdhdlr.client.SendMessageEvent(br.Room, event.EventMessage, dispatch)

	if br.Room == cmdhdlr.eventsChannel && cmdhdlr.eventsToStdout {
		fmt.Println("Event: ", br.Msg)
	}
}

// Helperfuncs for sending messages
func formatMessageForPing(cmdhdlr *CommandHandler, targetuser id.UserID) string {
	resp, err := cmdhdlr.client.GetDisplayName(targetuser)
	if err != nil {
		botNotifyEventsChannel(cmdhdlr, "FormatMessageForPing: "+err.Error())
		return ""
	}
	
	return "<a href=\"https://matrix.to/#/" + targetuser.String() + "\">" + resp.DisplayName + "</a>"
}


func botReplyMsg(cmdhdlr *CommandHandler, targetroom id.RoomID, msg string) id.EventID {
	resp, err := cmdhdlr.client.SendNotice(targetroom, msg)
	if err != nil {
		botNotifyEventsChannel(cmdhdlr, err.Error())
		return id.EventID("0")
	}
	return resp.EventID
}

func botReplyMsgEdited(cmdhdlr *CommandHandler, targetroom id.RoomID, msg string, edittarged id.EventID) {
	editedmsg := event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    msg,
	}
	editedmsg.SetEdit(edittarged)
	_, err := cmdhdlr.client.SendMessageEvent(targetroom, event.EventMessage, editedmsg)
	if err != nil {
		botNotifyEventsChannel(cmdhdlr, err.Error())
	}
}

func botNotifyEventsChannel(cmdhdlr *CommandHandler, msg string) {
	if cmdhdlr.eventsChannel != id.RoomID("0") {
		_, err := cmdhdlr.client.SendNotice(cmdhdlr.eventsChannel, msg)
		if err != nil && !cmdhdlr.eventsToStdout {
			fmt.Println("Event: ", msg) // Fallback
		}
	}
	if cmdhdlr.eventsToStdout {
		fmt.Println("Event: ", msg)
	}
}


// Archive just in case
/*
func botReplyMsgPinged(cmdhdlr *CommandHandler, targetuser id.UserID, targetroom id.RoomID, msg1 string, msg2 string) {
	reply := msg1 + targetuser.String() + msg2
	var replyformatted string

	pingstring := FormatMessageForPing(cmdhdlr, targetuser)
	if pingstring == "" {
		replyformatted = reply
	} else {
		replyformatted = msg1 + pingstring + msg2
	}

	BotReplyMsgFormatted(cmdhdlr, targetroom, reply, replyformatted)
}
*/

/*
func botReplyMsgWithReply(cmdhdlr *CommandHandler, targetroom id.RoomID, text string, replyto *event.Event) (*mautrix.RespSendEvent, error) {
	dispatch := event.MessageEventContent{
		MsgType:       event.MsgNotice,
		Body:          text,
	}
	dispatch.SetReply(replyto)
	return cmdhdlr.client.SendMessageEvent(targetroom, event.EventMessage, dispatch)
}
*/

/*
func botReplyMsgFormatted(cmdhdlr *CommandHandler, targetroom id.RoomID, msg string, msgformatted string) {
	_, err := cmdhdlr.client.SendMessageEvent(targetroom, event.EventMessage, &event.MessageEventContent{
		MsgType:       event.MsgNotice,
		Body:          msg,
		Format:        "org.matrix.custom.html",
		FormattedBody: msgformatted,
	})
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "BotReplyMsgFormatted Error 1:"+err.Error())
	}
}
*/

/*
func BotReplyMsgWithReply(cmdhdlr *CommandHandler, targetroom id.RoomID, text string, formatted string, replyto *event.Event) (*mautrix.RespSendEvent, error) {
	dispatch := event.MessageEventContent{
		MsgType:       event.MsgNotice,
		Body:          text,
		Format:        "org.matrix.custom.html",
		FormattedBody: formatted,
	}
	dispatch.SetReply(replyto)
	return cmdhdlr.client.SendMessageEvent(targetroom, event.EventMessage, dispatch)
}
*/

// Notes
/* // Edit a message
	//edits act on the original eventid, not the one returned from subsequent edits
	if strings.HasPrefix(messagebody, "edit") {
		msg := event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    "Before Message Edit",
		}
		eventid, _ := cmdhdlr.client.SendMessageEvent(evt.RoomID, event.EventMessage, msg)
		time.Sleep(1 * time.Second)

		msg = event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    "After Message Edit 1",
		}
		msg.SetEdit(eventid.EventID)
		_, _ = cmdhdlr.client.SendMessageEvent(evt.RoomID, event.EventMessage, msg)
		time.Sleep(1 * time.Second)

		msg = event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    "After Message Edit 2",
		}
		msg.SetEdit(eventid.EventID)
		_, _ = cmdhdlr.client.SendMessageEvent(evt.RoomID, event.EventMessage, msg)
	}
*/