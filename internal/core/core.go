package core

import "C"
import (
	"context"
	"errors"
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/config"
	"github.com/jialuohu/curlyraft/internal/clog"
	"github.com/jialuohu/curlyraft/internal/proto/raftcomm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type clientConnInfo struct {
	client raftcomm.RaftCommunicationClient
	conn   *grpc.ClientConn
	closer func(conn *grpc.ClientConn)
}

type RaftCore struct {
	raftcomm.UnimplementedRaftCommunicationServer
	raftcomm.UnimplementedHealthCheckServer
	Info      nodeInfo
	Peers     []nodeInfo
	peerConns map[string]*clientConnInfo

	node       *node
	grpcServer *grpc.Server
	rg         *RaftGateway

	mu            sync.Mutex
	receivedVotes uint16
	quorumSize    uint16
}

func NewRaftCore(cfg config.NodeCfg, sm curlyraft.StateMachine) *RaftCore {
	log.Printf("%s Creating new raft core\n", clog.CGreenRc("NewRaftCore"))
	peerStrParse := func(peerStr string) []nodeInfo {
		out := make([]nodeInfo, 0)
		peerStrList := strings.Split(peerStr, ",")
		for _, ps := range peerStrList {
			peerParts := strings.Split(ps, "/")
			out = append(out, nodeInfo{
				id:      peerParts[0],
				netAddr: peerParts[1],
			})
		}
		return out
	}

	rc := &RaftCore{
		Info: nodeInfo{
			id:      cfg.Id,
			netAddr: cfg.NetAddr,
		},
		Peers:         peerStrParse(cfg.Peers),
		node:          newNode(cfg.StorageDir, sm),
		receivedVotes: 0,
		quorumSize:    0,
		peerConns:     make(map[string]*clientConnInfo, len(cfg.Peers)),
	}

	clusterSize := uint16(len(rc.Peers) + 1)
	rc.quorumSize = (clusterSize / 2) + 1 // assumed clusterSize is odd

	log.Printf("%s Start building up client pool\n", clog.CGreenRc("NewRaftCore"))
	for _, p := range rc.Peers {
		client, conn, closer, err := rc.buildClient(p.netAddr)
		if err != nil {
			log.Fatalf("%s Failed to build grpc client: %v\n", clog.CRedRc("NewRaftCore"), err)
		}
		rc.peerConns[p.id] = &clientConnInfo{
			client: client,
			conn:   conn,
			closer: closer,
		}
	}
	log.Printf("%s Created new raft core\n", clog.CGreenRc("NewRaftCore"))
	return rc
}

func (rc *RaftCore) Start() error {
	log.Printf("%s Start listening on %s\n", clog.CGreenRc("Start"), rc.Info.netAddr)
	lis, err := net.Listen("tcp", rc.Info.netAddr)
	if err != nil {
		log.Printf("%s Failed to listen on %s: %v", clog.CRedRc("Start"), rc.Info.netAddr, err)
		return err
	}
	grpcServer := grpc.NewServer()
	raftcomm.RegisterRaftCommunicationServer(grpcServer, rc)
	raftcomm.RegisterHealthCheckServer(grpcServer, rc)
	rc.grpcServer = grpcServer
	go func() {
		log.Printf("%s The gRPC Server Start serving...", clog.CGreenRc("Start"))
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("%s The gRPC Server serve error: %v", clog.CRedRc("Start"), err)
		}
	}()
	return nil
}

func (rc *RaftCore) Run() error {
	conn, err := grpc.NewClient(rc.Info.netAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("%s Failed to build up grpc client: %v\n", clog.CRedRc("Run"), err)
		return err
	}
	client := raftcomm.NewHealthCheckClient(conn)
	for {
		res, err := client.Check(context.Background(), &raftcomm.HealthCheckRequest{Service: "run"})
		if err != nil {
			log.Printf("%s Failed to got message back: %v\n", clog.CRedRc("Run"), err)
			return err
		}
		status := res.GetStatus()
		if status == 1 {
			break
		}
		log.Printf("%s Waiting for grpc server startup...status:%d\n", clog.CGreenRc("Run"), status)
		time.Sleep(time.Millisecond * 100)
	}

	log.Printf("%s Start running rc.roleLoop()\n", clog.CGreenRc("Run"))
	return rc.roleLoop()
}

func (rc *RaftCore) Stop() error {
	log.Printf("%s Stop running raft core\n", clog.CGreenRc("Stop"))
	rc.node.stopNode()
	log.Printf("%s Stopped running raft node\n", clog.CGreenRc("Stop"))
	if rc.rg != nil {
		rc.rg.stop()
		log.Printf("%s Stopped raft gateway\n", clog.CGreenRc("Stop"))
	}
	if rc.grpcServer != nil {
		rc.grpcServer.GracefulStop()
		log.Printf("%s Stopped grpcServer\n", clog.CGreenRc("Stop"))
	}
	return nil
}
