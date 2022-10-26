package main

import (
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
	if len(rawuri) <= 0 {
		return nil, ErrNotARoomID
	}

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