package main

import (
	"testing"
	"bytes"
)


func TestAccountStoreSingle(t *testing.T) {
	var tests = []struct {
		inkey 	string
		invalue string
	}{
		{"sneed", "chuck"},
		{"lenny", "carl"},
		{"linux", "windows"},
	}

	for _, tt := range tests {
		tt := tt //you need this

		testname := "Test for " + tt.inkey
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			accountstore := NewDefaultAccountStorage("TT", true)

			err := accountstore.StoreData(nil, tt.inkey, &tt.invalue)
			if err != nil {
				t.Fatalf("failed to store data: "+err.Error())

			}

			var retrieve string
			err = accountstore.FetchData(nil, tt.inkey, &retrieve)
			if err != nil {
				t.Fatalf("failed to get data: "+err.Error())
			}

			if retrieve != tt.invalue {
				t.Fatalf("expected >%s< got >%s<", tt.invalue, retrieve)
			}
		})
	}
}

func TestAccountStoreMulti(t *testing.T) {
	var tests = []struct {
		inkey 	string
		invalue string
	}{
		{"sneed2", "chuck2"},
		{"lenny2", "carl2"},
		{"linux2", "windows2"},
	}
	accountstore := NewDefaultAccountStorage("TT2", true)

	for _, tt := range tests {
		testname := "Test for " + tt.inkey
		t.Run(testname, func(t *testing.T) {
			err := accountstore.StoreData(nil, tt.inkey, &tt.invalue)
			if err != nil {
				t.Fatalf("failed to store data: "+err.Error())

			}

			var retrieve string
			err = accountstore.FetchData(nil, tt.inkey, &retrieve)
			if err != nil {
				t.Fatalf("failed to get data: "+err.Error())
			}

			if retrieve != tt.invalue {
				t.Fatalf("expected >%s< got >%s<", tt.invalue, retrieve)
			}
		})
	}
}


func FuzzAccountStore(f *testing.F) {
	for _, seed := range [][]byte{[]byte("sneed"), []byte("chuck"), []byte("lenny"), []byte("carl"), []byte("linux"), []byte("windows")} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, invalue []byte) {
		accountstore := NewDefaultAccountStorage("FF", true)

		err := accountstore.StoreData(nil, "inkey", &invalue)
		if err != nil {
			t.Fatalf("%v -- failed to store data: %s", invalue, err.Error())

		}

		var retrieve []byte
		err = accountstore.FetchData(nil, "inkey", &retrieve)
		if err != nil {
			t.Fatalf("%v -- failed to get data: %s", invalue, err.Error())
		}

		if !bytes.Equal(retrieve, invalue) {
			t.Fatalf("expected >%v< got >%v<", invalue, retrieve)
		}
	})
}