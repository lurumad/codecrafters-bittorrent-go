package main

import (
	"encoding/json"
	"fmt"
	"os"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded := NewBencode().decode(bencodedValue)
		if decoded.err != nil {
			fmt.Println(decoded.err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded.value)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		file := os.Args[2]
		parsed := NewParser().parse(NewBencode(), file)
		fmt.Println("Tracker URL: " + parsed.metainfo.announce)
		fmt.Println("Length:", parsed.metainfo.info.length)

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
