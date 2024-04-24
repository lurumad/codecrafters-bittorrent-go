package main

import (
	"encoding/hex"
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
		hasher := NewBencodeSHA1Hasher()
		parsed := NewTorrentFileParser().parse(bencode, file)
		bencoded := bencode.encode(parsed.metainfo.info)
		hash := hasher.hash([]byte(bencoded.value))
		fmt.Println("Tracker URL: " + parsed.metainfo.announce)
		fmt.Println("Length:", parsed.metainfo.info["length"])
		fmt.Println("Info Hash:", hash)
		fmt.Println("Piece Length:", parsed.metainfo.info["piece length"])
		fmt.Println("Piece Hashes:")
		pieces := parsed.metainfo.info["pieces"].(string)
		for len(pieces) > 0 {
			piece := pieces[:20]
			fmt.Println(hex.EncodeToString([]byte(piece)))
			pieces = pieces[20:]
		}
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
