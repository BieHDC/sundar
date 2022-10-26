package main

import (
	"flag"
	"fmt"
	"encoding/json"
	"encoding/csv"
	"os"
	"strings"
	"strconv"
	"io"
)

type Echo struct {
	Callsign string
	Powerlevel int
	Message string
}

func main() {
	csvin := flag.String("in", "", "the csv pointing to the echo commands")
	jsonout := flag.String("out", "", "the json pointing to the output file to be fed into the bot")
	flag.Parse()
	if *csvin == "" || *jsonout == "" {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	csvraw, err := os.ReadFile(*csvin)
	if err != nil {
		panic(err)
	}
	csvparsed := csv.NewReader(strings.NewReader(string(csvraw)))

	stringconverter := strings.NewReplacer("??NL??", "\n")
	var echos []Echo
	for {
		record, err := csvparsed.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		pl, err := strconv.Atoi(record[1])
		if err != nil {
			panic(err)
		}

		echos = append(echos, Echo{Callsign: record[0], Powerlevel: pl, Message: stringconverter.Replace(record[2])})
	}

	bytes, err := json.Marshal(echos)
	if err != nil {
		panic(err)
	}

	os.WriteFile(*jsonout, bytes, 0600)
}