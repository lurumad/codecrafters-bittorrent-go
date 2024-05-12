package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"os"
)

var (
	ErrInvalidTorrentFile = errors.New("invalid torrent file")
	ErrInvalidMetainfo    = errors.New("invalid metainfo")
)

type TorrentParser struct {
	bencode *Bencode
}

type Info struct {
	Length      int
	Name        string
	PieceLength int
	Hash        []byte
	Pieces      [][]byte
}

type Metainfo struct {
	Announce string
	Info     Info
}

type Torrent struct {
	Metainfo *Metainfo
	Err      error
}

func NewTorrentParser(bencode *Bencode) *TorrentParser {
	return &TorrentParser{
		bencode: bencode,
	}
}

func (torrent *Torrent) ContainsPiece(index int) bool {
	return index >= 0 && index < len(torrent.Metainfo.Info.Pieces)
}

func (torrentFile *TorrentParser) Parse(filename string) *Torrent {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return &Torrent{
			Metainfo: nil,
			Err:      ErrInvalidTorrentFile,
		}
	}
	bencode := torrentFile.bencode
	decode := bencode.Decode(string(fileContents))
	if decode.err != nil {
		log.Println(decode.err)
		return &Torrent{
			Metainfo: nil,
			Err:      err,
		}
	}
	metainfo, ok := decode.value.(map[string]interface{})
	if !ok {
		fmt.Println("metainfo is invalid")
		return &Torrent{
			Metainfo: nil,
			Err:      ErrInvalidMetainfo,
		}
	}
	info, ok := metainfo["info"].(map[string]interface{})
	if !ok {
		fmt.Println("info is invalid")
		return &Torrent{
			Metainfo: nil,
			Err:      ErrInvalidMetainfo,
		}
	}
	length, ok := info["length"].(int)
	if !ok {
		fmt.Println("info.length is invalid")
		return &Torrent{
			Metainfo: nil,
			Err:      ErrInvalidMetainfo,
		}
	}
	pieceLength, ok := info["piece length"].(int)
	if !ok {
		fmt.Println("piece length is invalid")
		return &Torrent{
			Metainfo: nil,
			Err:      ErrInvalidMetainfo,
		}
	}
	hash := torrentFile.hash(bencode.encode(info))
	return &Torrent{
		Metainfo: &Metainfo{
			Announce: metainfo["announce"].(string),
			Info: Info{
				Length:      length,
				Name:        info["name"].(string),
				PieceLength: pieceLength,
				Hash:        hash,
				Pieces:      torrentFile.pieces(info),
			},
		},
		Err: nil,
	}
}

func (torrentFile *TorrentParser) hash(encode BencodeEncoded) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(encode.value))
	return hasher.Sum(nil)
}

func (torrentFile *TorrentParser) pieces(info map[string]interface{}) [][]byte {
	response := make([][]byte, 0)
	pieces := info["pieces"].(string)
	for len(pieces) > 0 {
		response = append(response, []byte(pieces[:20]))
		pieces = pieces[20:]
	}
	return response
}
