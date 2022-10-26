package main

import (
	"testing"
)



func TestSanitiseParseCommandWithArguments(t *testing.T) {
	var tests = []struct {
		name 				string
		stringin 			string
		prefix 				string
		arg_count 			int
		arg_values 			[]string
	}{
		{"nothing", 			"", 			"", 		0, []string{}},
		{"simple", 				"simple", 		"",			1, []string{"simple"}},
		{"two words", 			"two words", 	"",			2, []string{"two", "words"}},
		{"simple prefix", 		"++simple", 	"++",		1, []string{"simple"}},
		{"2 word prefix",		"~~word1 word2", 	"~~",		2, []string{"word1", "word2"}},
		{"both word prefixed",	"~~word3 ~~word4", 	"~~",		2, []string{"word3", "~~word4"}},
		{"multiple spaces",		"in the  past    spaces \tcaused confusion", 	"",		6, []string{"in", "the", "past", "spaces", "\tcaused", "confusion"}},
		{"newlines and spaces",	"\n trimmed hopefully \n", 	"", 2, []string{"trimmed", "hopefully"}},
		{"spaces 2",			"   tabs > spaces  \t ", 	"", 3, []string{"tabs", ">", "spaces"}},
	}

	for _, tt := range tests {
		tt := tt //you need this

		testname := "Test for >>" + tt.name + "<<"
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			values, count := sanitiseParseCommandWithArguments(tt.stringin, tt.prefix)

			if count != tt.arg_count {
				t.Fatalf("expected >%d<, got >%d< || exptected:>%#v< actual:>%#v<", tt.arg_count, count, tt.arg_values, values)
			}

			for i, v := range values {
				if v != tt.arg_values[i] {
					t.Errorf("mismatch at postition >%d<, expected >%s<, got >%s<", i, tt.arg_values[i], v)
				}
			}
		})
	}
}


func FuzzSanitiseParseCommandWithArguments(f *testing.F) {
	for _, seed := range []string{"command1", "command 2", "command three drei", "xx yy zz vv", "       ", "\n\n\n\n"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, stringin string) {
		arg_values, arg_count := sanitiseParseCommandWithArguments(stringin, "")
		_, _ = arg_values, arg_count

		t.Logf("\n\t%#v\n\t%#v", arg_values, stringin)
	})
}