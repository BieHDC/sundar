package main

import (
	id "maunium.net/go/mautrix/id"

	"testing"
)


func TestEchoCommands(t *testing.T) {
	type echotest struct {
		callsign 				string
		shouldsucceed 			bool
		expecteddesc 			string
		output 					string
	}

	var tests = []echotest{
		{"sneed", 		true,	"feeds and seeds", 															"feeds and seeds"},
		{"sneed", 		false, 	"should not have desc", 													"feeds and seeds 2"},
		{"chuck", 		true, 	"the previous owner", 														"the previous owner"},
		{"longlong", 	true, 	"if we get well over 70 chars in the message, the description should...", 	"if we get well over 70 chars in the message, the description should be truncated and get three dots added"},
	}

	cmdhdlr := NewCommandHandler(nil, "##", id.UserID("0"), id.RoomID("0"), false, "testing", true, "")

	t.Logf("adding the echos to the register")
	// Add all the echos and validate them
	for _, tt := range tests {
		tt := tt //you need this

		success := emitCommand(cmdhdlr, cmdhdlr.echoregister, Echo{Callsign: tt.callsign, Powerlevel: 0, Message: tt.output}, "testing")
		if success != tt.shouldsucceed {
			t.Fatalf("expected >%t<, got >%t<", tt.shouldsucceed, success)
		}

		// Skip the further validation below if we wanted to fail
		if !success && !tt.shouldsucceed {
			continue
		}

		command, exists := cmdhdlr.allcommands[tt.callsign]
		if !exists {
			t.Fatalf("expected echo command >%s< to exist, but it doesnt", tt.callsign)
		}

		if command.Description != tt.expecteddesc {
			t.Errorf("expected the description to be >%s<, but got >%s< instead", tt.expecteddesc, command.Description)
		}
	}
	
	t.Logf("appending an echo command that does not exist")
	// Append one command that does not exist
	tests = append(tests, echotest{callsign: "nadaexista", shouldsucceed: true, expecteddesc: "none", output: "Echo called >nadaexista< does not exist."})


	t.Logf("calling all the echo commands")
	// Call all the echos and check their output
	for _, tt := range tests {
		tt := tt //you need this

		testname := "Test for >>" + tt.callsign + "<<"
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			if !tt.shouldsucceed {
				// Should be skipped
				return
			}

			ca := CommandArgs{cmdhdlr, id.RoomID("0"), id.UserID("0"), 1, []string{tt.callsign}, id.RoomID("0"), nil, cmdhdlr.allcommands[tt.callsign]}
			botreply := HandleBotEchoMessage(ca)

			if botreply.Msg != tt.output {
				t.Errorf("expected >%s<, got >%s<", tt.output, botreply.Msg)
			}
		})
	}
}