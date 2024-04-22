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
		bencode := NewBencode()
		parsed := NewParser().parse(bencode, file)
		bencoded := bencode.encode(parsed.metainfo.info)
		hash := NewBencodeSHA1Hasher().hash(bencoded.value)
		fmt.Println("Tracker URL: " + parsed.metainfo.announce)
		fmt.Println("Length:", parsed.metainfo.info["length"])
		fmt.Println("Info Hash:", hash)

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
