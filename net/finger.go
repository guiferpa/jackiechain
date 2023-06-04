package net

type Finger struct {
	Addr string
}

func NewFinger(addr string) Finger {
	return Finger{addr}
}
