package core

import (
	"github.com/jialuohu/curlyraft/internal/clog"
	"github.com/jialuohu/curlyraft/internal/proto/gateway"
	"google.golang.org/grpc"
	"log"
)

type RaftGateway struct {
	gateway.UnimplementedRaftGatewayServer
	gateway.UnimplementedHealthCheckServer
	rc         *RaftCore
	grpcServer *grpc.Server
}

func NewRaftGateway(rc *RaftCore, grpcServer *grpc.Server) *RaftGateway {
	return &RaftGateway{rc: rc, grpcServer: grpcServer}
}

func (rg *RaftGateway) stop() {
	if rg.grpcServer != nil {
		rg.grpcServer.GracefulStop()
	}
	log.Printf("%s Stopped raft gateway gRPC server\n", clog.CGreenRg("stop"))
}
