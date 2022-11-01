package main

import (
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"net/url"
	"strings"
)

func geturlorpanic(unslaveprovider string) string {
	if unslaveprovider == "" {
		return ""
	}

	// Fixme: this thing is mostly just broken, i cant even
	// get the thing to panic. So eventually at some point
	// this thing should be taken care of.
	parsed, err := url.Parse(unslaveprovider)
	if err != nil {
		panic("broken unslaveprovider 1: "+err.Error())
	}

	// This is absolutely ridicilous, if your url is "a.b.c",
	// the "a.b.c" ends up in the Path, if you have a Scheme existing,
	// it is put in the Host.
	if parsed.Host != "" {
		return parsed.Host
	} else if parsed.Host == "" && parsed.Path != "" {
		return parsed.Path
	} else {
		panic("broken unslaveprovider 2: "+err.Error())
	}
}

func replaceshortsandunslaveHandler(unslaveprovider string) FunctionRegisterPrototype {
	unslaver := geturlorpanic(unslaveprovider)

	return func(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) BotReply {
		results := BotPrintNothing()
		for i := 0; i < argc; i++ {
			urlp, err := url.Parse(argv[i])
			if err != nil {
				continue
			}

			if strings.Contains(urlp.Host, "youtube.") || strings.Contains(urlp.Host, "youtu.be") {
				if strings.HasPrefix(urlp.Path, "/shorts/") {
					urlsplitted := strings.SplitN(urlp.Path, "/shorts/", 2)
					urlp.Path = "/watch"
					urlp.RawQuery = "v=" + urlsplitted[1]
					if unslaver != "" {
						unslavedurl := *urlp
						unslavedurl.Host = unslaver
						newentry := BotPrintSimple(room, "Fixed youtube link: " + urlp.String() + "\nUnslaved youtube link: " + unslavedurl.String()).WithNoBridge()
						BotPrintAppend(&results, &newentry)
					} else {
						newentry := BotPrintSimple(room, "Fixed youtube link: " + urlp.String()).WithNoBridge()
						BotPrintAppend(&results, &newentry)
					}
				} else {
					if unslaver != "" {
						unslavedurl := *urlp
						unslavedurl.Host = unslaver
						newentry := BotPrintSimple(room, "Unslaved youtube link: " + unslavedurl.String()).WithNoBridge()
						BotPrintAppend(&results, &newentry)
					}
				}
			}
		}
		return results
	}
}

func replacetwitterlinksHandler(unslaveprovider string) FunctionRegisterPrototype {
	unslaver := geturlorpanic(unslaveprovider)

	return func(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) BotReply {
		results := BotPrintNothing()
		if unslaver != "" {
			for i := 0; i < argc; i++ {
				urlp, err := url.Parse(argv[i])
				if err != nil {
					continue
				}

				if strings.Contains(urlp.Host, "twitter.") {
					urlp.Host = unslaver
					newentry := BotPrintSimple(room, "Fixed twitter link: " + urlp.String()).WithNoBridge()
					BotPrintAppend(&results, &newentry)
				}
			}
		}
		return results
	}
}


func ddgmeHandler(cmdhdlr *CommandHandler, searchproviderurl string, trigger string) FunctionRegisterPrototype {
	preparsed, err := url.Parse(searchproviderurl)
	if err != nil {
		panic(err)
	}
	whistle := trigger

	// If we request a string of text
	cmdhdlr.AddCommand(whistle, "Search the arguments or the message you replied to", "<reply>OR<question>", "Productivity", CommandAnyone, 
		func(ca CommandArgs) BotReply {
			if ca.argc < 2 {
				return BotPrintSimple(ca.statusroom, ca.self.Usage)
			}

			query := strings.Join(ca.argv[1:], " ")
			newurl := *preparsed
			newurl.RawQuery = newurl.RawQuery + url.QueryEscape(query)

			return BotPrintSimple(ca.room, "Let me search that for you: " + newurl.String()).WithReply(ca.evt)
		})

	// If we search for someone else
	return func(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, evt *event.Event) BotReply {
		if !strings.HasSuffix(evt.Content.AsMessage().Body, cmdhdlr.prefix+whistle) {
			return BotPrintNothing()
		}

		relation := evt.Content.AsMessage().GetRelatesTo()
		if relation == nil {
			return BotPrintNothing()
		}

		targetevent, err := cmdhdlr.client.GetEvent(room, relation.GetReplyTo())
		if err != nil {
			return BotPrintNothing()
		}

		err = targetevent.Content.ParseRaw(event.EventMessage)
		if err != nil {
			return BotPrintNothing()
		}
		
		newurl := *preparsed
		newurl.RawQuery = newurl.RawQuery + url.QueryEscape(targetevent.Content.AsMessage().Body)

		return BotPrintSimple(room, "Let me search that for you: " + newurl.String()).WithReply(targetevent)
	}
}
