package main

import (
	_ "embed"
	"encoding/xml"
	"fmt"
	"strings"
	"os"
	"bytes"
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

//go:embed bugcheck.xml
var BugCheckFile []byte

type ErrorCodesHresult struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Hresult"`
}

//go:embed hresult.xml
var HresultFile []byte

type ErrorCodesMmresult struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Mmresult"`
}

//go:embed mmresult.xml
var wmMmresult []byte

type ErrorCodesNtstatus struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Ntstatus"`
}

//go:embed ntstatus.xml
var wmNtstatus []byte

type ErrorCodesWinerror struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"Winerror"`
}

//go:embed winerror.xml
var wmWinerror []byte

type ErrorCodesWindowMessage struct {
	XMLName   xml.Name    //`xml:any`
	ErrorCode []ErrorCode `xml:"WindowMessage"`
}

//go:embed wm.xml
var WindowMessageFile []byte

// Return type
type ErrorCodesTargets struct {
	Text string
	Type string
}

func main() {
	allerrors := make(map[string][]ErrorCodesTargets)

	var errorsBugCheck ErrorCodesBugCheck
	err := xml.Unmarshal(BugCheckFile, &errorsBugCheck)
	if err != nil {
		return //err, "1"
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
		return //err, "2"
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
		return //err, "3"
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
		return //err, "4"
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
		return //err, "5"
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
		return //err, "6"
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

	var final bytes.Buffer

	final.WriteString("//THIS FILE IS AUTO GENERATED BY 'error_codes/generate_codes.go'\n\n")
	final.WriteString("package main\n\n")
	final.WriteString("type ErrorCodesTargets struct {\n	Text string\n	Type string\n}\n\n")
	final.WriteString("var WinErrorCodes map[string][]ErrorCodesTargets = map[string][]ErrorCodesTargets{\n")
	for key, values := range allerrors {
		final.WriteString("\t\""+key+"\": []ErrorCodesTargets{\n")
		for _, value := range values {
			final.WriteString("\t\t{Text: \""+value.Text+"\", Type: \""+value.Type+"\"},\n")
		}
		final.WriteString("\t},\n")
	}
	final.WriteString("}\n\n")

	os.WriteFile("../custom_error_codes_handler_errors.go", final.Bytes(), 0600)
}