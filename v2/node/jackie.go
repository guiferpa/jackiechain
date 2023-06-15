package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/guiferpa/jackiechain/blockchain"
)

const (
	JACKIE_DOWNLOAD_BLOACKCHAIN_OK = "DOWNLOAD_BLOCKCHAIN_OK"
	JACKIE_DOWNLOAD_BLOACKCHAIN    = "DOWNLOAD_BLOCKCHAIN"
	JACKIE_TX_APPROBATION_FAIL     = "TX_APPROBATION_FAIL"
	JACKIE_TX_APPROBATION_OK       = "TX_APPROBATION_OK"
	JACKIE_TX_APPROBATION          = "TX_APPROBATION"
	JACKIE_CONNECT_OK              = "CONNECT_OK"
	JACKIE_CONNECT_LOOPBACK        = "CONNECT_LOOPBACK"
	JACKIE_CONNECT                 = "CONNECT"
	JACKIE_DISCONNECT              = "DISCONNECT"
)

func Send(addr, action string, args []string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("JACKIE %s %s", action, strings.Join(args, " "))
	if _, err := conn.Write([]byte(message)); err != nil {
		return err
	}

	return nil
}

func PeerConnectRequest(id, addr, to string) error {
	args := []string{id, addr}
	return Send(to, JACKIE_CONNECT, args)
}

func PeerConnectLoopback(peer Peer, port, to string) error {
	conn, err := net.Dial("udp", to)
	if err != nil {
		return err
	}

	ip := conn.LocalAddr().(*net.UDPAddr).IP
	addr := fmt.Sprintf("%s:%s", ip, port)
	args := []string{peer.GetID(), addr}
	return Send(to, JACKIE_CONNECT_LOOPBACK, args)
}

func PeerDisconnect(peer Peer, to string) error {
	args := []string{peer.GetID()}
	return Send(to, JACKIE_DISCONNECT, args)
}

func TxApprobationRequest(peer Peer, tx blockchain.Transaction, to string) error {
	bs, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	args := []string{peer.GetID(), base64.StdEncoding.EncodeToString(bs)}
	return Send(to, JACKIE_TX_APPROBATION, args)
}

func TxApprobationOK(peer Peer, txhash, to string) error {
	args := []string{peer.GetID(), txhash}
	return Send(to, JACKIE_TX_APPROBATION_OK, args)
}

func DownloadBlockchainRequest(id string, to string) error {
	args := []string{id}
	return Send(to, JACKIE_DOWNLOAD_BLOACKCHAIN, args)
}

func DownloadBlockchainOK(id, chain, to string) error {
	args := []string{id, chain}
	return Send(to, JACKIE_DOWNLOAD_BLOACKCHAIN_OK, args)
}
