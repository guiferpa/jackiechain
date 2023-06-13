package tcp

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const (
	JACKIE_SYNC_UPTIME_OK          = "SYNC_UPTIME_OK"
	JACKIE_SYNC_UPTIME             = "SYNC_UPTIME"
	JACKIE_DOWNLOAD_BLOACKCHAIN_OK = "DOWNLOAD_BLOCKCHAIN_OK"
	JACKIE_DOWNLOAD_BLOACKCHAIN    = "DOWNLOAD_BLOCKCHAIN"
	JACKIE_TX_APPROBATION_FAIL     = "TX_APPROBATION_FAIL"
	JACKIE_TX_APPROBATION_OK       = "TX_APPROBATION_OK"
	JACKIE_TX_APPROBATION          = "TX_APPROBATION"
	JACKIE_CONNECT_OK              = "CONNECT_OK"
	JACKIE_CONNECT_LOOPBACK        = "CONNECT_LOOPBACK"
	JACKIE_CONNECT                 = "CONNECT"
	JACKIE_DISCONNECT              = "DISCONNECT"
	JACKIE_MESSAGE                 = "MESSAGE"
)

type PeerJackie struct {
	ID   string `json:"id"`
	Host string `json:"host"`
	Port string `json:"port"`
}

func (pj *PeerJackie) GetAddr() string {
	return fmt.Sprintf("%s:%s", pj.Host, pj.Port)
}

func NewPeerJackie(id, addr string) *PeerJackie {
	host := ""
	port := ""

	addrspl := strings.Split(addr, ":")
	if len(addrspl) == 0 {
		panic("invalid peer jackie address")
	}

	if len(addrspl) == 1 {
		port = addrspl[0]
	}

	if len(addrspl) == 2 {
		host = addrspl[0]
		port = addrspl[1]
	}

	return &PeerJackie{
		ID:   id,
		Host: host,
		Port: port,
	}
}

func ParseJackieRequest(b *bytes.Buffer) (string, []string, error) {
	message := strings.Split(b.String(), " ")

	if len(message) < 2 {
		return "", nil, errors.New("incorrect jackie length protocol")
	}

	proto := message[0]
	act := message[1]

	if proto == "JACKIE" {
		return act, message[2:], nil
	}

	return "", nil, errors.New("invalid jackie protocol")
}
