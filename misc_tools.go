package main

import (
	"fmt"
	"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strings"
	"errors"
)

// Helperfuncs for roomid stuff
var (
	ErrNotARawRoomID 	= errors.New("rawuri must start with '!' to be a rawuri")
	ErrNotARoomID 		= errors.New("string seems to not be an roomid of any kind")
)

func rawURIStringToURI(rawuri string) (*id.MatrixURI, error) {
	if rawuri[0] != '!' {
		return nil, ErrNotARoomID
	}
	// "!theidstringoflenxx:reactos.org?via=somenonsense"
	idpart := strings.SplitN(rawuri, "?", 2)[0]
	// "!theidstringoflenxx:reactos.org"
	idpart = strings.TrimPrefix(idpart, "!")
	// "theidstringoflenxx:reactos.org"

	var parseduri id.MatrixURI
	parseduri = id.MatrixURI{
		Sigil1: '!',
		MXID1:  idpart,
	}
	return &parseduri, nil
}

func parseRoomID(rawuri string) (*id.MatrixURI, error) {
	parsedMatrix, err := id.ParseMatrixURIOrMatrixToURL(rawuri)
	if err == nil && parsedMatrix != nil {
		return parsedMatrix, nil
	} else {
		// maybe its a raw uri
		rawuriparsed, erruri := rawURIStringToURI(rawuri)
		if erruri == nil && rawuriparsed != nil {
			parsedMatrix, err = id.ParseMatrixURIOrMatrixToURL(rawuriparsed.String())
			if err == nil && parsedMatrix != nil {
				return parsedMatrix, nil
			}
		} else {
			return nil, erruri
		}
	}

	return nil, ErrNotARoomID
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


// Helperfuncs for sending messages
func FormatMessageForPing(cmdhdlr *CommandHandler, targetuser id.UserID) string {
	resp, err := cmdhdlr.client.GetDisplayName(targetuser)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, "HandleHello Error 1:"+err.Error())
		return ""
	} else {
		return "<a href=\"https://matrix.to/#/" + targetuser.String() + "\">" + resp.DisplayName + "</a>"
	}
}

func BotReplyMsgPinged(cmdhdlr *CommandHandler, targetuser id.UserID, targetroom id.RoomID, msg1 string, msg2 string) {
	reply := msg1 + targetuser.String() + msg2
	replyformatted := ""

	pingstring := FormatMessageForPing(cmdhdlr, targetuser)
	if pingstring == "" {
		replyformatted += reply
	} else {
		replyformatted += msg1 + pingstring + msg2
	}

	BotReplyMsgFormatted(cmdhdlr, targetroom, reply, replyformatted)
}

func BotReplyMsgFormatted(cmdhdlr *CommandHandler, targetroom id.RoomID, msg string, msgformatted string) {
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

func BotReplyMsg(cmdhdlr *CommandHandler, targetroom id.RoomID, msg string) {
	_, err := cmdhdlr.client.SendNotice(targetroom, msg)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, err.Error())
	}
}

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

func BotNotifyEventsChannel(cmdhdlr *CommandHandler, msg string) {
	if cmdhdlr.eventsChannel != "" {
		_, err := cmdhdlr.client.SendNotice(cmdhdlr.eventsChannel, msg)
		if err != nil && !cmdhdlr.eventsToStdout {
			fmt.Println("Event: ", msg) // Fallback
		}
	}
	if cmdhdlr.eventsToStdout {
		fmt.Println("Event: ", msg)
	}
}

func BotReplyMsgRet(cmdhdlr *CommandHandler, targetroom id.RoomID, msg string) id.EventID {
	resp, err := cmdhdlr.client.SendNotice(targetroom, msg)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, err.Error())
		return id.EventID(0)
	}
	return resp.EventID
}

func BotReplyMsgEdited(cmdhdlr *CommandHandler, targetroom id.RoomID, msg string, edittarged id.EventID) {
	editedmsg := event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    msg,
	}
	editedmsg.SetEdit(edittarged)
	_, err := cmdhdlr.client.SendMessageEvent(targetroom, event.EventMessage, editedmsg)
	if err != nil {
		BotNotifyEventsChannel(cmdhdlr, err.Error())
	}
}

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