package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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
		parse := NewTorrentParser(bencode).Parse(file)
		if parse.Err != nil {
			log.Fatal(parse.Err)
		}
		fmt.Println("Tracker URL: " + parse.Metainfo.Announce)
		fmt.Println("Length:", parse.Metainfo.Info.Length)
		fmt.Println("Info Hash:", hex.EncodeToString(parse.Metainfo.Info.Hash))
		fmt.Println("Piece Length:", parse.Metainfo.Info.PieceLength)
		fmt.Println("Piece Hashes:")
		for _, value := range parse.Metainfo.Info.Pieces {
			fmt.Println(hex.EncodeToString(value))
		}
	} else if command == "peers" {
		file := os.Args[2]
		bencode := NewBencode()
		torrentFile := NewTorrentParser(bencode).Parse(file)
		if torrentFile.Err != nil {
			log.Fatal(torrentFile.Err)
		}
		peers, err := NewTorrentClient(bencode).Peers(torrentFile, "00112233445566778899")
		if err != nil {
			log.Fatal(err)
		}
		for _, peer := range peers {
			fmt.Printf("%v:%d\n", peer.IP, peer.Port)
		}
	} else if command == "handshake" {
		file := os.Args[2]
		address := os.Args[3]
		bencode := NewBencode()
		torrent := NewTorrentParser(bencode).Parse(file)
		if torrent.Err != nil {
			log.Fatal(torrent.Err)
		}
		handshake := NewTorrentClient(bencode).Handshake(address, torrent.Metainfo.Info.Hash)
		if handshake.Err != nil {
			log.Fatal(handshake.Err)
		}
		fmt.Printf("Peer ID: %v\n", handshake.PeerId)
	} else if command == "download_piece" {
		output := os.Args[3]
		file := os.Args[4]
		piece, err := strconv.Atoi(os.Args[5])
		if err != nil {
			log.Fatal(err)
		}
		bencode := NewBencode()
		torrent := NewTorrentParser(bencode).Parse(file)
		if torrent.Err != nil {
			log.Fatal(torrent.Err)
		}
		client := NewTorrentClient(bencode)
		err = client.DownloadPiece(&PieceRequest{
			Piece:   piece,
			PeerId:  "00112233445566778899",
			Torrent: torrent,
			Output:  output,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Piece %d downloaded to %v.\n", piece, output)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
