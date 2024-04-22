package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

var (
	ErrInvalidTorrentFile = errors.New("invalid torrent file")
	ErrInvalidMetainfo    = errors.New("invalid metainfo")
)

type TorrentFileParser struct {
}

type Metainfo struct {
	announce string
	info     map[string]interface{}
}

type Parsed struct {
	metainfo *Metainfo
	err      error
}

func NewParser() *TorrentFileParser {
	return &TorrentFileParser{}
}

func (p *TorrentFileParser) parse(b *Bencode, filename string) *Parsed {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return &Parsed{
			metainfo: nil,
			err:      ErrInvalidTorrentFile,
		}
	}

	bencode := string(fileContents)
	decode := b.decode(bencode)
	if decode.err != nil {
		log.Println(decode.err)
		return &Parsed{
			metainfo: nil,
			err:      err,
		}
	}

	metainfo, ok := decode.value.(map[string]interface{})
	if !ok {
		fmt.Println("metainfo is invalid")
		return &Parsed{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	info, ok := metainfo["info"].(map[string]interface{})
	if !ok {
		fmt.Println("info is invalid")
		return &Parsed{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	length, ok := info["length"].(int)
	if !ok {
		fmt.Println("info.length is invalid")
		return &Parsed{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	pieceLength, ok := info["piece length"].(int)
	if !ok {
		fmt.Println("piece length is invalid")
		return &Parsed{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	response := Metainfo{
		announce: metainfo["announce"].(string),
		info: map[string]interface{}{
			"length":       length,
			"name":         info["name"].(string),
			"piece length": pieceLength,
			"pieces":       info["pieces"].(string),
		},
	}
	return &Parsed{
		metainfo: &response,
		err:      nil,
	}
}
