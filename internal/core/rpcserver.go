package core

import (
	"context"
	raftcomm "github.com/jialuohu/curlyraft/internal/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

type RpcServer struct {
	raftcomm.UnimplementedRaftCommunicationServer
}

func NewRpcServer(netAddr string) {
	log.Printf("[NewRpcServer] Starting rpc server\n")
	lis, err := net.Listen("tcp", netAddr)
	if err != nil {
		log.Printf("[NewRpcServer] Error while listening on %s: %v\n", netAddr, err)
		return
	}
	server := &RpcServer{}
	grpcServer := grpc.NewServer()
	raftcomm.RegisterRaftCommunicationServer(grpcServer, server)
	if err := grpcServer.Serve(lis); err != nil {
		log.Printf("[NewRpcServer] Error while serving: %v\n", err)
	}
}

func (r RpcServer) AppendEntries(ctx context.Context, request *raftcomm.AppendEntriesRequest) (*raftcomm.AppendEntriesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (r RpcServer) RequestVote(ctx context.Context, request *raftcomm.RequestVoteRequest) (*raftcomm.RequestVoteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (r RpcServer) InstallSnapShot(ctx context.Context, request *raftcomm.InstallSnapshotRequest) (*raftcomm.InstallSnapshotResponse, error) {
	//TODO implement me
	panic("implement me")
}
