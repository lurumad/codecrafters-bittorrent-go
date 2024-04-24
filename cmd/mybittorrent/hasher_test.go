package main

import (
	"testing"
)

func TestHashInfoDictionary(t *testing.T) {
	want := "1cad4a486798d952614c394eb15e75bec587fd08"
	bencode := NewBencode()
	parsed := NewTorrentFileParser().parse(bencode, "../../sample.torrent")
	if parsed.err != nil {
		t.Fatal(parsed.err)
	}
	bencoded := bencode.encode(parsed.metainfo.info)

	hash := NewBencodeSHA1Hasher().hash([]byte(bencoded.value))

	if hash != want {
		t.Errorf("wrong hash - want %v, got %v", want, hash)
	}
}
