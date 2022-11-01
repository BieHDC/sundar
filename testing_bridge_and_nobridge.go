package main

import(

)

func HandleBridgeTesting(ca CommandArgs) BotReply {
	bridged := BotPrintSimple(ca.room, "This message should be bridged")
	notbridged := BotPrintSimple(ca.room, "This message may contain gamer words that should not be bridged").WithNoBridge()
	BotPrintAppend(&bridged, &notbridged)
	return bridged
}