package peer

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/guiferpa/jackiechain/blockchain"
	"github.com/guiferpa/jackiechain/logger"
	protogreeter "github.com/guiferpa/jackiechain/proto/greeter"
	protonet "github.com/guiferpa/jackiechain/proto/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type ID string

type Remote string

type ProtoClients struct {
	net protonet.NetClient
}

func GetProtoClients(remote Remote) (*ProtoClients, error) {
	conn, err := grpc.NewClient(string(remote), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	protoClients := &ProtoClients{
		net: protonet.NewNetClient(conn),
	}
	return protoClients, nil
}

type Peer struct {
	ID            ID
	IP            []byte
	Port          int
	PeerRemoteMap map[ID]Remote
	Blockchain    *blockchain.Blockchain
	protogreeter.UnimplementedGreeterServer
	protonet.UnimplementedNetServer
}

func (p *Peer) ReachOut(ctx context.Context, pr *protogreeter.PingRequest) (*protogreeter.PongResponse, error) {
	logger.Yellow(fmt.Sprintf("Ping from agent %s", pr.Aid))
	return &protogreeter.PongResponse{Pid: string(p.ID)}, nil
}

func (p *Peer) Connect(ctx context.Context, cr *protonet.ConnectRequest) (*protonet.ConnectResponse, error) {
	pctx, ok := peer.FromContext(ctx)
	if !ok {
		return nil, nil
	}
	logger.Yellow(fmt.Sprintf("Connection request from peer %s", cr.Pid))
	if len(p.PeerRemoteMap) > 0 {
		var wg sync.WaitGroup
		for id, premote := range p.PeerRemoteMap {
			wg.Add(1)
			go func(id string, remote Remote) {
				defer wg.Done()
				clients, err := GetProtoClients(remote)
				if err != nil {
					logger.Red(err.Error())
					return
				}
				logger.Yellow(fmt.Sprintf("Send connection to peer %s", id))
				scr := &protonet.SendConnectionRequest{Pid: cr.Pid, Remote: cr.Remote}
				_, err = clients.net.SendConnection(ctx, scr)
				if err != nil {
					logger.Red(err.Error())
					return
				}
			}(string(id), premote)
		}
		wg.Wait()
	}
	host, _, err := net.SplitHostPort(pctx.Addr.String())
	if err != nil {
		logger.Red(err.Error())
		return nil, nil
	}
	_, port, err := net.SplitHostPort(pctx.LocalAddr.String())
	if err != nil {
		logger.Red(err.Error())
		return nil, nil
	}
	p.PeerRemoteMap[ID(cr.Pid)] = Remote(fmt.Sprintf("%v:%v", host, port))
	return &protonet.ConnectResponse{Pid: string(p.ID), Status: uint32(0)}, nil
}

func (p *Peer) SendConnection(ctx context.Context, scr *protonet.SendConnectionRequest) (*protonet.SendConnectionResponse, error) {
	logger.Yellow(fmt.Sprintf("Received connection about peer %s", scr.Pid))
	p.PeerRemoteMap[ID(scr.Pid)] = Remote(scr.Remote)
	return nil, nil
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
	logger.Yellow(fmt.Sprintf("Try connect IP(%v), Port(%v) to peer in network", p.IP, p.Port))
	cr := &protonet.ConnectRequest{
		Pid:    string(p.ID),
		Remote: "",
	}
	resp, err := netclient.Connect(context.Background(), cr)
	if err != nil {
		return err
	}
	if resp.Status != 0 {
		return fmt.Errorf("TryConnect method failured with status equals %v", resp.Status)
	}
	logger.Yellow(fmt.Sprintf("Connection successful with peer %s", resp.Pid))
	return nil
}

func (p *Peer) Serve(listener net.Listener, nodeRemote string, serving chan struct{}, cherr chan error) {
	if nodeRemote != "" {
		addr, err := net.ResolveUDPAddr("udp", nodeRemote)
		if err != nil {
			panic(err)
		}
		p.IP = addr.IP
		p.Port = addr.Port
	}

	s := grpc.NewServer()
	protogreeter.RegisterGreeterServer(s, p)
	protonet.RegisterNetServer(s, p)
	serving <- struct{}{}
	if err := s.Serve(listener); err != nil {
		cherr <- err
	}
}

func New(id ID, bc *blockchain.Blockchain) *Peer {
	return &Peer{ID: id, Blockchain: bc, PeerRemoteMap: make(map[ID]Remote, 0)}
}
