package main

import (
	"errors"
	"testing"
)

func TestErrParseTorrentFile(t *testing.T) {
	parsed := NewParser().parse(NewBencode(), "invalid filename")
	if !errors.Is(parsed.err, ErrInvalidTorrentFile) {
		t.Errorf("expected ErrInvalidTorrentFile - got: %v", parsed.err)
	}

	if parsed.metainfo != nil {
		t.Errorf("error metainfo should be nil - got: %v", parsed.metainfo)
	}
}

func TestParse(t *testing.T) {
	parsed := NewParser().parse(NewBencode(), "../../sample.torrent")
	if parsed.err != nil {
		t.Fatal(parsed.err)
	}

	if parsed.metainfo == nil {
		t.Errorf("error metainfo should not be empty")
	}

	if parsed.metainfo.announce != "http://bittorrent-test-tracker.codecrafters.io/announce" {
		t.Errorf("announce bad result - want %v, got %v", "http://bittorrent-test-tracker.codecrafters.io/announce", parsed.metainfo.announce)
	}

	if parsed.metainfo.info.length != 92063 {
		t.Errorf("length bad result - want %d, got %d", 92063, parsed.metainfo.info.length)
	}

	if parsed.metainfo.info.pieceLength != 32768 {
		t.Errorf("piece length bad result - want %d, got %d", 32768, parsed.metainfo.info.pieceLength)
	}

	if parsed.metainfo.info.name != "sample.txt" {
		t.Errorf("name bad result - want %v, got %v", "sample.txt", parsed.metainfo.info.name)
	}
}
