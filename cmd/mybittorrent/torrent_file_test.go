package main

import (
	"encoding/hex"
	"errors"
	"testing"
)

func TestErrParseTorrentFile(t *testing.T) {
	parsed := NewTorrentParser().Parse("invalid filename", NewBencode())
	if !errors.Is(parsed.Err, ErrInvalidTorrentFile) {
		t.Errorf("expected ErrInvalidTorrentFile - got: %v", parsed.Err)
	}

	if parsed.Metainfo != nil {
		t.Errorf("error metainfo should be nil - got: %v", parsed.Metainfo)
	}
}

func TestParse(t *testing.T) {
	type piecesTestCase struct {
		want string
	}

	parsed := NewTorrentParser().Parse("../../sample.torrent", NewBencode())
	if parsed.Err != nil {
		t.Fatal(parsed.Err)
	}

	metainfo := parsed.Metainfo
	info := metainfo.Info

	if metainfo == nil {
		t.Errorf("error metainfo should not be empty")
	}

	if metainfo.Announce != "http://bittorrent-test-tracker.codecrafters.io/announce" {
		t.Errorf("announce bad result - want %v, got %v", "http://bittorrent-test-tracker.codecrafters.io/announce", metainfo.Announce)
	}

	if info.Length != 92063 {
		t.Errorf("length bad result - want %d, got %d", 92063, info.Length)
	}

	if info.PieceLength != 32768 {
		t.Errorf("piece length bad result - want %d, got %d", 32768, info.PieceLength)
	}

	if info.Name != "sample.txt" {
		t.Errorf("name bad result - want %v, got %v", "sample.txt", info.Name)
	}

	if hex.EncodeToString(info.Hash) != "d69f91e6b2ae4c542468d1073a71d4ea13879a7f" {
		t.Errorf("wrong hash - want d69f91e6b2ae4c542468d1073a71d4ea13879a7f, got %v", hex.EncodeToString(info.Hash))
	}

	if len(info.Pieces) != 3 {
		t.Errorf("wrong pieces length - want 3, got %v", len(info.Pieces))
	}

	for index, tc := range []piecesTestCase{
		{want: "e876f67a2a8886e8f36b136726c30fa29703022d"},
		{want: "6e2275e604a0766656736e81ff10b55204ad8d35"},
		{want: "f00d937a0213df1982bc8d097227ad9e909acc17"},
	} {
		if hex.EncodeToString(info.Pieces[index]) != tc.want {
			t.Errorf("wrong piece hash - want %v, got %v", tc.want, hex.EncodeToString(info.Pieces[index]))
		}
	}
}
