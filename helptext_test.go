package main

import (
	id "maunium.net/go/mautrix/id"

	"testing"
)


func TestGenerateHelp(t *testing.T) {
	type echotest struct {
		callsign 	string
		desc 		string
		usage 		string
		category 	string
		pl 			int
	}

	var tests = []echotest{
		{"sneed", 	"feeds and seeds", 		"",			"fun",		0},
		{"chuck", 	"the previous owner", 	"<userid>",	"serious",	5},
		{"long", 	"long complex128", 		"[length]",	"joke",		33},
	}

	cmdhdlr := NewCommandHandler(nil, "##", id.UserID("0"), id.RoomID("0"), false, "testing", true, "")

	t.Logf("adding example functions")
	// Add all the echos and validate them
	for _, tt := range tests {
		tt := tt //you need this

		cmdhdlr.AddCommand(tt.callsign, tt.desc, tt.usage, tt.category, tt.pl, nil)

		_, exists := cmdhdlr.allcommands[tt.callsign]
		if !exists {
			t.Fatalf("expected command >%s< to exist, but it doesnt", tt.callsign)
		}
	}
	

	t.Logf("getting the whole helptext")
	text, formatted := cmdhdlr.generateHelp(id.RoomID("0"), 100)

	expected_text := "```\nAvailable Commands for you in this room:\nfun:\n\tsneed -> feeds and seeds\n\njoke:\n\tlong -> long complex128 -> Required Power Level: 33\n\nserious:\n\tchuck -> the previous owner -> Required Power Level: 5\n\n```"
	expected_formatted := "<pre><code>Available Commands for you in this room:\nfun:\n\tsneed -&gt; feeds and seeds\n\njoke:\n\tlong -&gt; long complex128 -&gt; Required Power Level: 33\n\nserious:\n\tchuck -&gt; the previous owner -&gt; Required Power Level: 5\n</code></pre>\n"
	
	if text != expected_text {
		t.Errorf("expected >>\n%q\n<<, got >>\n%q\n<<", expected_text, text)
	}
	if formatted != expected_formatted {
		t.Errorf("expected >>\n%q\n<<, got >>\n%q\n<<", expected_formatted, formatted)
	}
}