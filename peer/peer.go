package peer

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/logger"
	protogreeter "github.com/guiferpa/jackiechain/proto/greeter"
	protonet "github.com/guiferpa/jackiechain/proto/net"
	"google.golang.org/grpc"
)

type ID string

type Remote string

type protoClients struct {
	net protonet.NetClient
}

type Peer struct {
	ID            ID
	IP            []byte
	Port          int
	PeerRemoteMap map[ID]Remote
	Blockchain    *blockchain.Blockchain
	protoClients  protoClients
	protogreeter.UnimplementedGreeterServer
	protonet.UnimplementedNetServer
}

func (p *Peer) ReachOut(ctx context.Context, pr *protogreeter.PingRequest) (*protogreeter.PongResponse, error) {
	logger.Yellow(fmt.Sprintf("Ping from agent %s", pr.Aid))
	return &protogreeter.PongResponse{Pid: string(p.ID)}, nil
}

func (p *Peer) Connect(ctx context.Context, cr *protonet.ConnectRequest) (*protonet.ConnectResponse, error) {
	logger.Yellow(fmt.Sprintf("Connection request from peer %s", cr.Pid))
	p.PeerRemoteMap[ID(cr.Pid)] = Remote(cr.Remote)
	return &protonet.ConnectResponse{Pid: string(p.ID), Status: uint32(0)}, nil
}

func (p *Peer) SetBuildBlockInterval(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			bh, err := blockchain.BuildBlock(p.Blockchain)
			if err != nil {
				logger.Red(err.Error())
				continue
			}
			logger.Magenta(fmt.Sprintf("Block %s was built", bh))
		}
	}
}

func (p *Peer) TryConnect(conn grpc.ClientConnInterface) error {
	netclient := protonet.NewNetClient(conn)
	cr := &protonet.ConnectRequest{
		Pid:    string(p.ID),
		Remote: fmt.Sprintf("%v:%v", p.IP, p.Port),
	}
	resp, err := netclient.Connect(context.Background(), cr)
	if err != nil {
		return err
	}
	if resp.Status != 0 {
		return fmt.Errorf("TryConnect method failured with status equals %v", resp.Status)
	}
	p.PeerRemoteMap[ID(resp.Pid)] = Remote(fmt.Sprintf("%v:%v", p.IP, p.Port))
	logger.Yellow(fmt.Sprintf("Connection successful with peer %s", resp.Pid))
	return nil
}

func (p *Peer) Serve(listener net.Listener, serving chan struct{}, cherr chan error) {
	dconn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	dconn.Close()
	laddr := dconn.LocalAddr().(*net.UDPAddr)
	p.IP = laddr.IP
	p.Port = laddr.Port

	s := grpc.NewServer()
	protogreeter.RegisterGreeterServer(s, p)
	protonet.RegisterNetServer(s, p)
	serving <- struct{}{}
	if err := s.Serve(listener); err != nil {
		cherr <- err
	}
}

func New(id ID, bc *blockchain.Blockchain) *Peer {
	return &Peer{ID: id, Blockchain: bc, PeerRemoteMap: make(map[ID]Remote, 0), protoClients: protoClients{}}
}
