package main

import (
	"testing"
)

func TestHashInfoDictionary(t *testing.T) {
	want := "d69f91e6b2ae4c542468d1073a71d4ea13879a7f"
	bencode := NewBencode()
	parsed := NewParser().parse(bencode, "../../sample.torrent")
	if parsed.err != nil {
		t.Fatal(parsed.err)
	}
	bencoded := bencode.encode(parsed.metainfo.info)

	hash := NewBencodeSHA1Hasher().hash(bencoded.value)
	
	if hash != want {
		t.Errorf("wrong hash - want %v, got %v", want, hash)
	}
}
