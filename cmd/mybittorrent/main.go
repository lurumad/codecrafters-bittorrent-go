package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded := NewBencode().Decode(bencodedValue)
		if decoded.err != nil {
			fmt.Println(decoded.err)
			return
		}
		jsonOutput, _ := json.Marshal(decoded.value)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		file := os.Args[2]
		bencode := NewBencode()
		parse := NewParseTorrentFile().Parse(bencode, file)
		if parse.err != nil {
			log.Fatal(parse.err)
		}
		fmt.Println("Tracker URL: " + parse.metainfo.announce)
		fmt.Println("Length:", parse.metainfo.info.length)
		fmt.Println("Info Hash:", hex.EncodeToString(parse.metainfo.info.hash))
		fmt.Println("Piece Length:", parse.metainfo.info.pieceLength)
		fmt.Println("Piece Hashes:")
		for _, value := range parse.metainfo.info.pieces {
			fmt.Println(hex.EncodeToString(value))
		}
	} else if command == "peers" {
		file := os.Args[2]
		bencode := NewBencode()
		parse := NewParseTorrentFile().Parse(bencode, file)
		if parse.err != nil {
			log.Fatal(parse.err)
		}
		trackerInfo := NewTracker().GetPeers(parse.metainfo, bencode)
		if trackerInfo.err != nil {
			log.Fatal(trackerInfo.err)
		}
		for _, peer := range trackerInfo.peers {
			fmt.Printf("%v:%d\n", peer.ip, peer.port)
		}
	} else if command == "handshake" {
		file := os.Args[2]
		address := os.Args[3]
		bencode := NewBencode()
		parse := NewParseTorrentFile().Parse(bencode, file)
		if parse.err != nil {
			log.Fatal(parse.err)
		}
		response := NewPeer().Handshake(&HandshakeRequest{
			address:  address,
			infoHash: parse.metainfo.info.hash,
		})
		if response.err != nil {
			log.Fatal(response.err)
		}
		fmt.Printf("Peer ID: %v\n", response.peerId)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
