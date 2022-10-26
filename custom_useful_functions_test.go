package main

import (
	id "maunium.net/go/mautrix/id"

	"testing"
)


func TestYoutubeURLfixWithUnslave(t *testing.T) {
	var tests = []struct {
		urlin 			string
		fixedout 		string
	}{
		{"this https://www.youtube.com/watch?v=fk_zqDGmssI", 		"Unslaved youtube link: https://piped.based.quest/watch?v=fk_zqDGmssI"},
		{"https://www.youtube.com/shorts/4FUYkvel5Uo video", 		"Fixed youtube link: https://www.youtube.com/watch?v=4FUYkvel5Uo\nUnslaved youtube link: https://piped.based.quest/watch?v=4FUYkvel5Uo"},
		{"this https://www.youtube.com/watch?v=A5Qo9jDNNAw video",	"Unslaved youtube link: https://piped.based.quest/watch?v=A5Qo9jDNNAw"},
		{"does not contain a youtube link", 						""},
		{"https://youtube.com/watch/1ue74jzFxD4?feature=share", 	"Unslaved youtube link: https://piped.based.quest/watch/1ue74jzFxD4?feature=share"},
	}

	unslavers := []string{"https://piped.based.quest", "piped.based.quest"}

	for _, ttt := range unslavers {
		ttt := ttt //just in case

		linkfix := replaceshortsandunslaveHandler(ttt)
		for _, tt := range tests {
			tt := tt //you need this

			testname := "Test for >>" + tt.urlin + "<<"
			t.Run(testname, func(t *testing.T) {
				t.Parallel()

				newargv, newargc := sanitiseParseCommandWithArguments(tt.urlin, "")
				botreply := linkfix(nil, id.RoomID("0"), id.UserID("0"), newargc,  newargv, nil)
				parsedreply := func() string {
					finalstring := ""
					cursor := &botreply
					for {
						if cursor.Print {
							finalstring += cursor.Msg
						}
						if cursor.next != nil {
							cursor = cursor.next
						} else {
							break
						}
					}
					return finalstring
				}()

				if parsedreply != tt.fixedout {
					t.Fatalf("expected >%s< got >%s<", tt.fixedout, parsedreply)
				}
			})
		}
	}
}

func TestYoutubeURLfix(t *testing.T) {
	var tests = []struct {
		urlin 			string
		fixedout 		string
	}{
		{"this https://www.youtube.com/watch?v=fk_zqDGmssI", 		""},
		{"https://www.youtube.com/shorts/4FUYkvel5Uo video", 		"Fixed youtube link: https://www.youtube.com/watch?v=4FUYkvel5Uo"},
		{"this https://www.youtube.com/watch?v=A5Qo9jDNNAw video",	""},
		{"does not contain a youtube link", 						""},
		{"https://youtube.com/watch/1ue74jzFxD4?feature=share", 	""},
	}


	linkfix := replaceshortsandunslaveHandler("") // no unslaver
	for _, tt := range tests {
		tt := tt //you need this

		testname := "Test for >>" + tt.urlin + "<<"
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			newargv, newargc := sanitiseParseCommandWithArguments(tt.urlin, "")
			botreply := linkfix(nil, id.RoomID("0"), id.UserID("0"), newargc,  newargv, nil)
			parsedreply := func() string {
				finalstring := ""
				cursor := &botreply
				for {
					if cursor.Print {
						finalstring += cursor.Msg
					}
					if cursor.next != nil {
						cursor = cursor.next
					} else {
						break
					}
				}
				return finalstring
			}()

			if parsedreply != tt.fixedout {
				t.Fatalf("expected >%s< got >%s<", tt.fixedout, parsedreply)
			}
		})
	}
}


func TestTwitterURLfix(t *testing.T) {
	var tests = []struct {
		urlin 			string
		fixedout 		string
	}{
		{"follow https://twitter.com/reactos", 													"Fixed twitter link: https://nitter.based.quest/reactos"},
		{"check out https://twitter.com/reactos/status/1586357185180680192 its awesome", 		"Fixed twitter link: https://nitter.based.quest/reactos/status/1586357185180680192"},
		{"https://twitter.com/hashtag/ReactOS to find all of them",								"Fixed twitter link: https://nitter.based.quest/hashtag/ReactOS"},
	}


	linkfix := replacetwitterlinksHandler("nitter.based.quest")
	for _, tt := range tests {
		tt := tt //you need this

		testname := "Test for >>" + tt.urlin + "<<"
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			newargv, newargc := sanitiseParseCommandWithArguments(tt.urlin, "")
			botreply := linkfix(nil, id.RoomID("0"), id.UserID("0"), newargc,  newargv, nil)
			parsedreply := func() string {
				finalstring := ""
				cursor := &botreply
				for {
					if cursor.Print {
						finalstring += cursor.Msg
					}
					if cursor.next != nil {
						cursor = cursor.next
					} else {
						break
					}
				}
				return finalstring
			}()

			if parsedreply != tt.fixedout {
				t.Fatalf("expected >%s< got >%s<", tt.fixedout, parsedreply)
			}
		})
	}
}


func TestSearchProvider(t *testing.T) {
	var tests = []struct {
		term 			string
		finalquery 		string
	}{
		{"", 									"Usage: ##search <reply>OR<question>"},
		{"i am searching for this", 			"Let me search that for you: https://duckduckgo.com/?q=i+am+searching+for+this"},
		{"i might also search for that", 		"Let me search that for you: https://duckduckgo.com/?q=i+might+also+search+for+that"},
		{"https://twitter.com/hashtag/ReactOS",	"Let me search that for you: https://duckduckgo.com/?q=https%3A%2F%2Ftwitter.com%2Fhashtag%2FReactOS"},
	}
	searchurl := "https://duckduckgo.com/?q="
	searchcallsign := "search"

	cmdhdlr := NewCommandHandler(nil, "##", id.UserID("0"), id.RoomID("0"), false, "testing", true, "")

	ddgmeHandler(cmdhdlr, searchurl, searchcallsign) //we cant test the function part without a full mocker

	for _, tt := range tests {
		tt := tt //you need this

		testname := "Test for >>" + tt.term + "<<"
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			newargv, newargc := sanitiseParseCommandWithArguments(searchcallsign+" "+tt.term, "")
			ca := CommandArgs{cmdhdlr, id.RoomID("0"), id.UserID("0"), newargc, newargv, id.RoomID("0"), nil, cmdhdlr.allcommands[searchcallsign]}
			botreply := ca.self.Targetfunc(ca)

			if botreply.Msg != tt.finalquery {
				t.Fatalf("expected >%s< got >%s<", tt.finalquery, botreply.Msg)
			}
		})
	}
}