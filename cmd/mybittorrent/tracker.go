package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type Tracker struct{}

func NewTracker() *Tracker {
	return &Tracker{}
}

type TrackerResponse struct {
	peers []PeerResponse
	err   error
}

type PeerResponse struct {
	ip   string
	port uint16
}

func (*Tracker) GetPeers(metainfo *Metainfo, b *Bencode) TrackerResponse {
	url, err := trackerUrl(metainfo)
	if err != nil {
		return TrackerResponse{err: err}
	}
	body := makeRequest(url)
	decode := b.Decode(string(body))
	if decode.err != nil {
		return TrackerResponse{peers: nil, err: decode.err}
	}
	tracker := decode.value.(map[string]interface{})
	peers := []byte(tracker["peers"].(string))
	response := TrackerResponse{}
	for i := 0; i < len(peers); i = i + 6 {
		ip := fmt.Sprintf("%d.%d.%d.%d", peers[i], peers[i+1], peers[i+2], peers[i+3])
		port := binary.BigEndian.Uint16(peers[i+4 : i+6])
		response.peers = append(response.peers, PeerResponse{ip: ip, port: port})
	}
	return response
}

func makeRequest(trackerUrl *url.URL) []byte {
	request, err := http.NewRequest("GET", trackerUrl.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func trackerUrl(metainfo *Metainfo) (*url.URL, error) {
	baseUrl, err := url.Parse(metainfo.announce)
	if err != nil {
		log.Fatal(err)
	}
	params := url.Values{}
	params.Add("info_hash", string(metainfo.info.hash))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.Itoa(metainfo.info.length))
	params.Add("compact", "1")
	baseUrl.RawQuery = params.Encode()
	return baseUrl, nil
}
