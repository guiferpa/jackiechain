package net

import "github.com/guiferpa/jackchain/blockchain"

type Node struct {
	Addr  string
	Chain *blockchain.Chain
}

func NewNode(addr string, chain *blockchain.Chain) *Node {
	return &Node{Addr: addr, Chain: chain}
}
