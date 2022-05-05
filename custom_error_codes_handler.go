package main

import (
	_ "embed"
	"encoding/xml"
	"fmt"
	"strings"
	//"maunium.net/go/mautrix"
	event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
)

// Sources for the xml files:
// https://github.com/reactos/Message_Translator/tree/master/GUI/Resources

type ErrorCode struct {
	XMLName xml.Name //`xml:any`
	Text    string   `xml:"text,attr"`
	Value   string   `xml:"value,attr"`
}

type ErrorCodesBugCheck struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"BugCheck"`
}

//go:embed error_codes/bugcheck.xml
var BugCheckFile []byte

type ErrorCodesHresult struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Hresult"`
}

//go:embed error_codes/hresult.xml
var HresultFile []byte

type ErrorCodesMmresult struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Mmresult"`
}

//go:embed error_codes/mmresult.xml
var wmMmresult []byte

type ErrorCodesNtstatus struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Ntstatus"`
}

//go:embed error_codes/ntstatus.xml
var wmNtstatus []byte

type ErrorCodesWinerror struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Winerror"`
}

//go:embed error_codes/winerror.xml
var wmWinerror []byte

type ErrorCodesWindowMessage struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"WindowMessage"`
}

//go:embed error_codes/wm.xml
var WindowMessageFile []byte

// Return type
type ErrorCodesTargets struct {
	Text string
	Type string
}

var allerrors map[string][]ErrorCodesTargets

func LoadErrors() (error, string) {
	allerrors = make(map[string][]ErrorCodesTargets)

	var errorsBugCheck ErrorCodesBugCheck
	err := xml.Unmarshal(BugCheckFile, &errorsBugCheck)
	if err != nil {
		return err, "1"
	} else {
		for i := 0; i < len(errorsBugCheck.ErrorCode); i++ {
			key := strings.ToUpper(strings.TrimLeft(errorsBugCheck.ErrorCode[i].Value, "0"))
			appendee := ErrorCodesTargets{errorsBugCheck.ErrorCode[i].Text, errorsBugCheck.ErrorCode[i].XMLName.Local}
			allerrors[key] = append(allerrors[key], appendee)
		}
	}

	var errorsHresult ErrorCodesHresult
	err = xml.Unmarshal(HresultFile, &errorsHresult)
	if err != nil {
		return err, "2"
	} else {
		for i := 0; i < len(errorsHresult.ErrorCode); i++ {
			key := strings.ToUpper(strings.TrimLeft(errorsHresult.ErrorCode[i].Value, "0"))
			appendee := ErrorCodesTargets{errorsHresult.ErrorCode[i].Text, errorsHresult.ErrorCode[i].XMLName.Local}
			allerrors[key] = append(allerrors[key], appendee)
		}
	}

	var errorsMmresult ErrorCodesMmresult
	err = xml.Unmarshal(wmMmresult, &errorsMmresult)
	if err != nil {
		return err, "3"
	} else {
		for i := 0; i < len(errorsMmresult.ErrorCode); i++ {
			key := strings.ToUpper(strings.TrimLeft(errorsMmresult.ErrorCode[i].Value, "0"))
			appendee := ErrorCodesTargets{errorsMmresult.ErrorCode[i].Text, errorsMmresult.ErrorCode[i].XMLName.Local}
			allerrors[key] = append(allerrors[key], appendee)
		}
	}

	var errorsNtstatus ErrorCodesNtstatus
	err = xml.Unmarshal(wmNtstatus, &errorsNtstatus)
	if err != nil {
		return err, "4"
	} else {
		for i := 0; i < len(errorsNtstatus.ErrorCode); i++ {
			key := strings.ToUpper(strings.TrimLeft(errorsNtstatus.ErrorCode[i].Value, "0"))
			appendee := ErrorCodesTargets{errorsNtstatus.ErrorCode[i].Text, errorsNtstatus.ErrorCode[i].XMLName.Local}
			allerrors[key] = append(allerrors[key], appendee)
		}
	}

	// These 2 are a special case where the value is in base 10 but we need it in base 16
	// Using Sscan and Sprintf seems to be the least annoying
	var errorsWinerror ErrorCodesWinerror
	err = xml.Unmarshal(wmWinerror, &errorsWinerror)
	if err != nil {
		return err, "5"
	} else {
		for i := 0; i < len(errorsWinerror.ErrorCode); i++ {
			var key string
			if errorsWinerror.ErrorCode[i].Value == "0" { //hackaround, for some reason sscan and sprintf are not happy about this case, and directly setting key to "0" doesnt work ether? ok sure
				key = strings.ToUpper(strings.TrimLeft(errorsWinerror.ErrorCode[i].Value, "0"))
			} else {
				var decimal uint
				_, _ = fmt.Sscan(strings.ToUpper(strings.TrimLeft(errorsWinerror.ErrorCode[i].Value, "0")), &decimal)
				key = string(fmt.Sprintf("%X", decimal))
			}
			appendee := ErrorCodesTargets{errorsWinerror.ErrorCode[i].Text, errorsWinerror.ErrorCode[i].XMLName.Local}
			allerrors[key] = append(allerrors[key], appendee)
		}
	}

	var errorsWindowMessage ErrorCodesWindowMessage
	err = xml.Unmarshal(WindowMessageFile, &errorsWindowMessage)
	if err != nil {
		return err, "6"
	} else {
		for i := 0; i < len(errorsWindowMessage.ErrorCode); i++ {
			var key string
			if errorsWindowMessage.ErrorCode[i].Value == "0" { //hackaround, for some reason sscan and sprintf are not happy about this case, and directly setting key to "0" doesnt work ether? ok sure
				key = strings.ToUpper(strings.TrimLeft(errorsWindowMessage.ErrorCode[i].Value, "0"))
			} else {
				var decimal uint
				_, _ = fmt.Sscan(strings.ToUpper(strings.TrimLeft(errorsWindowMessage.ErrorCode[i].Value, "0")), &decimal)
				key = string(fmt.Sprintf("%X", decimal))
			}
			appendee := ErrorCodesTargets{errorsWindowMessage.ErrorCode[i].Text, errorsWindowMessage.ErrorCode[i].XMLName.Local}
			allerrors[key] = append(allerrors[key], appendee)
		}
	}

	return nil, "0"
}

func GetErrors(cmdhdlr *CommandHandler, parsederror string) string {
	if allerrors == nil {
		err, num := LoadErrors()
		if err != nil {
			BotNotifyEventsChannel(cmdhdlr, "LoadErrors Error "+num+":"+err.Error())
			return "Failed to load the Error Codes."
		}
	}
	reply := ""
	errors, exists := allerrors[parsederror]
	if exists {
		reply += "The following codes have been found:\n"
		for _, error := range errors {
			reply += "\t" + error.Type + " -> " + error.Text + "\n"
		}
	} else {
		reply += "Error Code not found: " + parsederror
	}

	return reply
}

func HandleErrorCodeRequest(cmdhdlr *CommandHandler, room id.RoomID, sender id.UserID, argc int, argv []string, statusroom id.RoomID, evt *event.Event) bool {
	if argc < 2 {
		cmdhdlr.internelPrintUsage(argv[0], room)
		return false
	} else {
		parsederrorcode := strings.TrimPrefix(argv[1], "0x")
		parsederrorcode =  strings.TrimLeft(parsederrorcode, "0")
		parsederrorcode =  strings.TrimRight(parsederrorcode, "\n")
		parsederrorcode =  strings.ToUpper(parsederrorcode)

		BotReplyMsg(cmdhdlr, room, GetErrors(cmdhdlr, parsederrorcode))
		return true
	}
}
