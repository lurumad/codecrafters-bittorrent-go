package main

import (
	"encoding/hex"
	"log"
	"net"
)

type Peer struct{}

func NewPeer() *Peer {
	return &Peer{}
}

type HandshakeRequest struct {
	address  string
	infoHash []byte
	err      error
}

type HandshakeResponse struct {
	peerId string
	err    error
}

func (p *Peer) Handshake(request *HandshakeRequest) *HandshakeResponse {
	connection, err := net.Dial(
		"tcp",
		request.address,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	handshake := handShakeMessage(request)
	_, err = connection.Write(handshake)
	if err != nil {
		log.Fatal(err)
	}
	buffer := make([]byte, 68)
	_, err = connection.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}
	return &HandshakeResponse{
		peerId: hex.EncodeToString(buffer[48:68]),
		err:    nil,
	}
}

func handShakeMessage(request *HandshakeRequest) []byte {
	var handshake []byte
	handshake = append(handshake, byte(19))
	handshake = append(handshake, []byte("BitTorrent protocol")...)
	handshake = append(handshake, make([]byte, 8)...)
	handshake = append(handshake, request.infoHash...)
	handshake = append(handshake, []byte("00112233445566778899")...)
	return handshake
}
