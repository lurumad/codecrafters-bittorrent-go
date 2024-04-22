package main

import (
	"crypto/sha1"
	"encoding/hex"
)

type BencodeSHA1Hasher struct{}

func NewBencodeSHA1Hasher() *BencodeSHA1Hasher {
	return &BencodeSHA1Hasher{}
}

func (h *BencodeSHA1Hasher) hash(info string) string {
	hasher := sha1.New()
	hasher.Write([]byte(info))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}
