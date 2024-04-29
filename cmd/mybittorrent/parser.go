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

type ParseTorrentFile struct {
}

type Info struct {
	length      int
	name        string
	pieceLength int
	hash        []byte
	pieces      [][]byte
}

type Metainfo struct {
	announce string
	info     Info
}

type ParseTorrentFileResult struct {
	metainfo *Metainfo
	err      error
}

func NewParseTorrentFile() *ParseTorrentFile {
	return &ParseTorrentFile{}
}

func (p *ParseTorrentFile) Parse(b *Bencode, filename string) *ParseTorrentFileResult {
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return &ParseTorrentFileResult{
			metainfo: nil,
			err:      ErrInvalidTorrentFile,
		}
	}

	bencode := string(fileContents)
	decode := b.Decode(bencode)
	if decode.err != nil {
		log.Println(decode.err)
		return &ParseTorrentFileResult{
			metainfo: nil,
			err:      err,
		}
	}

	metainfo, ok := decode.value.(map[string]interface{})
	if !ok {
		fmt.Println("metainfo is invalid")
		return &ParseTorrentFileResult{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	info, ok := metainfo["info"].(map[string]interface{})
	if !ok {
		fmt.Println("info is invalid")
		return &ParseTorrentFileResult{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	length, ok := info["length"].(int)
	if !ok {
		fmt.Println("info.length is invalid")
		return &ParseTorrentFileResult{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	pieceLength, ok := info["piece length"].(int)
	if !ok {
		fmt.Println("piece length is invalid")
		return &ParseTorrentFileResult{
			metainfo: nil,
			err:      ErrInvalidMetainfo,
		}
	}

	hash := p.hashInfo(b.encode(info))

	return &ParseTorrentFileResult{
		metainfo: &Metainfo{
			announce: metainfo["announce"].(string),
			info: Info{
				length:      length,
				name:        info["name"].(string),
				pieceLength: pieceLength,
				hash:        hash,
				pieces:      p.pieces(info),
			},
		},
		err: nil,
	}
}

func (p *ParseTorrentFile) hashInfo(encode BencodeEncoded) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(encode.value))
	return hasher.Sum(nil)
}

func (p *ParseTorrentFile) pieces(info map[string]interface{}) [][]byte {
	response := make([][]byte, 0)
	pieces := info["pieces"].(string)
	for len(pieces) > 0 {
		response = append(response, []byte(pieces[:20]))
		pieces = pieces[20:]
	}
	return response
}
