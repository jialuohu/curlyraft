package core

import (
	"context"
	raftcomm "github.com/jialuohu/curlyraft/internal/proto"
)

func (rc *RaftCore) AppendEntries(ctx context.Context, request *raftcomm.AppendEntriesRequest) (*raftcomm.AppendEntriesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (rc *RaftCore) RequestVote(ctx context.Context, request *raftcomm.RequestVoteRequest) (*raftcomm.RequestVoteResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (rc *RaftCore) InstallSnapShot(ctx context.Context, request *raftcomm.InstallSnapshotRequest) (*raftcomm.InstallSnapshotResponse, error) {
	//TODO implement me
	panic("implement me")
}
