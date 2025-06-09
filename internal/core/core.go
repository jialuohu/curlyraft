package core

import (
	"errors"
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/config"
	raftcomm "github.com/jialuohu/curlyraft/internal/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"strings"
	"sync"
)

type RaftCore struct {
	raftcomm.UnimplementedRaftCommunicationServer
	Info  nodeInfo
	Peers []nodeInfo

	node       *node
	grpcServer *grpc.Server

	mu            sync.Mutex
	receivedVotes uint16
	quorumSize    uint16
}

func NewRaftCore(cfg config.NodeCfg, sm curlyraft.StateMachine) *RaftCore {
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
	}

	clusterSize := uint16(len(rc.Peers) + 1)
	rc.quorumSize = (clusterSize / 2) + 1 // assumed clusterSize is odd
	return rc
}

func (rc *RaftCore) Start() error {
	log.Printf("[RaftCore] Start listening on %s\n", rc.Info.netAddr)
	lis, err := net.Listen("tcp", rc.Info.netAddr)
	if err != nil {
		log.Printf("[RaftCore] Failed to listen on %s: %v", rc.Info.netAddr, err)
		return err
	}

	grpcServer := grpc.NewServer()
	raftcomm.RegisterRaftCommunicationServer(grpcServer, rc)
	rc.grpcServer = grpcServer
	go func() {
		log.Println("[RpcServer] The gRPC Server Start serving...")
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("[RaftCore] The gRPC Server serve error: %v", err)
		}
	}()
	return nil
}

func (rc *RaftCore) Run() error {
	return rc.roleLoop()
}

func (rc *RaftCore) Stop() {
	rc.node.stopNode()
	if rc.grpcServer != nil {
		rc.grpcServer.GracefulStop()
	}
}
