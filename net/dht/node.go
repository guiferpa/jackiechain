package dht

import "github.com/guiferpa/jackchain/net"

type Node struct {
	*net.Node
	Sucessor    *net.Node
	FingerTable FingerTable
}

func (n *Node) Join(jn *net.Node) error {
	return nil
}

func NewNode() (*Node, error) {
	return nil, nil
}
