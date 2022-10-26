package main

import (
	"testing"
)


func TestWindowsErrorCodes(t *testing.T) {
	var tests = []struct {
		errorstring 	string
		expectedresult 	string
	}{
		{"21", 			"The following codes have been found:\n\tBugCheck -> QUOTA_UNDERFLOW\n\tMmresult -> UNDEFINED\n\tWinerror -> ERROR_LOCK_VIOLATION\n\tWindowMessage -> WM_MOUSEACTIVATE\n"},
		{"420", 		"The following codes have been found:\n\tWinerror -> ERROR_SERVICE_ALREADY_RUNNING\n"},
		{"FA", 			"The following codes have been found:\n\tBugCheck -> HTTP_DRIVER_CORRUPTED\n\tWindowMessage -> undefined_87\n"},
		{"1337", 		"Error Code not found: 1337"},
		{"nadaexista", 	"Error Code not found: NADAEXISTA"},
	}

	for _, tt := range tests {
		tt := tt //you need this

		testname := "Test for " + tt.errorstring
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			theerrors := sanitiseinputandgeterrors(tt.errorstring)

			if theerrors != tt.expectedresult {
				t.Fatalf("expected >%s< got >%s<", tt.expectedresult, theerrors)
			}
		})
	}
}


func FuzzWindowsErrorCodes(f *testing.F) {
	for _, seed := range []string{"21", "420", "FA", "1337", "nada", "6656876786972697687"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, invalue string) {
		sanitiseinputandgeterrors(invalue)
	})
}
