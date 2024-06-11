package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type TorrentClient struct {
	bencode    *Bencode
	httpClient *http.Client
	connection net.Conn
}

func NewTorrentClient(bencode *Bencode) *TorrentClient {
	return &TorrentClient{
		bencode:    bencode,
		httpClient: &http.Client{},
	}
}

type Peer struct {
	IP   string
	Port uint16
}

type Handshake struct {
	PeerId string
	Err    error
}

type PieceRequest struct {
	Address string
	Piece   int
	Torrent *Torrent
	Output  string
}

type DownloadRequest struct {
	Address string
	Torrent *Torrent
	Output  string
}

type PeerMessage struct {
	Id      int32
	Payload interface{}
}

type PiecePayload struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

type PieceBlockPayload struct {
	Index int32
	Begin int32
	Block []byte
}

type MessageType int32

const (
	Choke MessageType = iota
	Unchoke
	Interested
	NotInterested
	Have
	Bitfield
	Request
	Piece
	Cancel
)

const HandshakeMessageLen = 68
const BlockSize = 16 * 1024

func (peer *Peer) Address() string {
	return peer.IP + ":" + strconv.Itoa(int(peer.Port))
}

func (tc *TorrentClient) ConnectToPeer(address string) error {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	tc.connection = connection
	return nil
}

func (tc *TorrentClient) Peers(torrent *Torrent, peerId string) ([]Peer, error) {
	url, err := trackerUrl(torrent, peerId)
	if err != nil {
		return make([]Peer, 0), err
	}
	body, err := tc.trackerDoGet(url)
	if err != nil {
		return make([]Peer, 0), err
	}
	decode := tc.bencode.Decode(string(body))
	if decode.err != nil {
		return make([]Peer, 0), err
	}
	tracker := decode.value.(map[string]interface{})
	peers := []byte(tracker["peers"].(string))
	response := make([]Peer, 0)
	for i := 0; i < len(peers); i = i + 6 {
		ip := fmt.Sprintf("%d.%d.%d.%d", peers[i], peers[i+1], peers[i+2], peers[i+3])
		port := binary.BigEndian.Uint16(peers[i+4 : i+6])
		response = append(response, Peer{IP: ip, Port: port})
	}
	return response, nil
}

func trackerUrl(torrent *Torrent, peerId string) (*url.URL, error) {
	baseUrl, err := url.Parse(torrent.Metainfo.Announce)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("info_hash", string(torrent.Metainfo.Info.Hash))
	params.Add("peer_id", peerId)
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.Itoa(torrent.Metainfo.Info.Length))
	params.Add("compact", "1")
	baseUrl.RawQuery = params.Encode()
	return baseUrl, nil
}

func (tc *TorrentClient) trackerDoGet(url *url.URL) ([]byte, error) {
	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := tc.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (tc *TorrentClient) Handshake(torrent *Torrent, address string) *Handshake {
	var handshake []byte
	handshake = append(handshake, byte(19))
	handshake = append(handshake, []byte("BitTorrent protocol")...)
	handshake = append(handshake, make([]byte, 8)...)
	handshake = append(handshake, torrent.Metainfo.Info.Hash...)
	handshake = append(handshake, []byte("00112233445566778899")...)
	if tc.connection == nil {
		if err := tc.ConnectToPeer(address); err != nil {
			return &Handshake{Err: err}
		}
	}
	var _, err = tc.connection.Write(handshake)
	if err != nil {
		return &Handshake{Err: err}
	}
	buffer := make([]byte, HandshakeMessageLen)
	_, err = tc.connection.Read(buffer)
	if err != nil {
		return &Handshake{Err: err}
	}
	peerId := hex.EncodeToString(buffer[48:HandshakeMessageLen])
	return &Handshake{
		PeerId: peerId,
		Err:    nil,
	}
}

func (tc *TorrentClient) DownloadPiece(request *PieceRequest) ([]byte, error) {
	tc.ConnectToPeer(request.Address)
	defer tc.connection.Close()
	tc.Handshake(request.Torrent, request.Address)
	tc.waitForMessage(Bitfield, tc.connection)
	message := PeerMessage{
		Id:      int32(Interested),
		Payload: struct{}{},
	}
	tc.sendMessage(message, tc.connection)
	tc.waitForMessage(Unchoke, tc.connection)
	if !request.Torrent.ContainsPiece(request.Piece) {
		return nil, errors.New("info.pieces does not contain piece")
	}
	numberOfFullBlocks := request.pieceLength() / BlockSize
	lastBlockLength := request.pieceLength() % BlockSize
	var data []byte
	for blockNumber := 0; blockNumber < numberOfFullBlocks; blockNumber++ {
		buffer, err := tc.pieceBlock(request.Piece, blockNumber, BlockSize, tc.connection)
		if err != nil {
			return nil, err
		}
		data = append(data, buffer...)
	}
	if lastBlockLength > 0 {
		buffer, err := tc.pieceBlock(request.Piece, numberOfFullBlocks, lastBlockLength, tc.connection)
		if err != nil {
			return nil, err
		}
		data = append(data, buffer...)
	}
	os.WriteFile(request.Output, data, os.ModePerm)
	return data, nil
}

func (tc *TorrentClient) Download(request *DownloadRequest) error {
	var data []byte
	for piece := range request.Torrent.Metainfo.Info.Pieces {
		piece, err := tc.DownloadPiece(&PieceRequest{
			Address: request.Address,
			Piece:   piece,
			Torrent: request.Torrent,
			Output:  request.Output,
		})
		if err != nil {
			return err
		}
		data = append(data, piece...)
	}
	os.WriteFile(request.Output, data, os.ModePerm)
	return nil
}

func (tc *TorrentClient) pieceBlock(piece int, blockNumber int, blockLength int, connection net.Conn) ([]byte, error) {
	message := PeerMessage{
		Id: int32(Request),
		Payload: PiecePayload{
			Index:  uint32(piece),
			Begin:  uint32(blockNumber * BlockSize),
			Length: uint32(blockLength),
		},
	}
	tc.sendMessage(message, connection)
	buffer, err := tc.ReadPieceBlock(connection)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func (tc *TorrentClient) sendMessage(message PeerMessage, connection net.Conn) error {
	buffer, err := serialize(message)
	if err != nil {
		return err
	}
	_, err = connection.Write(buffer)
	if err != nil {
		return err
	}
	return nil
}

func (tc *TorrentClient) ReadPieceBlock(connection net.Conn) ([]byte, error) {
	buffer := make([]byte, 4)
	if _, err := connection.Read(buffer); err != nil {
		return nil, err
	}
	lengthPrefix := binary.BigEndian.Uint32(buffer)
	payloadBuffer := make([]byte, lengthPrefix)
	_, err := io.ReadFull(connection, payloadBuffer)
	if err != nil {
		return nil, err
	}
	message, err := deserialize(payloadBuffer)
	if err != nil {
		return nil, err
	}
	if payload, ok := message.Payload.(PieceBlockPayload); ok {
		return payload.Block, nil
	}
	return nil, errors.New("expected PieceBlockPayload")
}

func (tc *TorrentClient) waitForMessage(messageType MessageType, connection net.Conn) error {
	buffer := make([]byte, 4)
	if _, err := connection.Read(buffer); err != nil {
		return err
	}
	lengthPrefix := binary.BigEndian.Uint32(buffer)
	payloadBuf := make([]byte, lengthPrefix)
	if _, err := connection.Read(payloadBuf); err != nil {
		return err
	}
	message, err := deserialize(payloadBuf)
	if err != nil {
		return err
	}
	if message.Id != int32(messageType) {
		return err
	}
	return nil
}

func (request *PieceRequest) pieceLength() int {
	rest := request.Torrent.Metainfo.Info.Length - (request.Torrent.Metainfo.Info.PieceLength * request.Piece)
	if rest >= request.Torrent.Metainfo.Info.PieceLength {
		return request.Torrent.Metainfo.Info.PieceLength
	}
	return rest
}

func serialize(message PeerMessage) ([]byte, error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, message.Payload)
	data := buf.Bytes()
	buffer := make([]byte, uint32(len(data))+5)
	binary.BigEndian.PutUint32(buffer[:4], uint32(len(data)+1))
	buffer[4] = byte(message.Id)
	copy(buffer[5:], data)
	return buffer, nil
}

func deserialize(buffer []byte) (PeerMessage, error) {
	id := int32(buffer[0])
	var payload interface{}
	switch id {
	case int32(Piece):
		payload = PieceBlockPayload{
			Index: 0,
			Begin: 0,
			Block: buffer[9:],
		}
	}
	return PeerMessage{
		Id:      id,
		Payload: payload,
	}, nil
}
