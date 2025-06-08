package core

import (
	"context"
	raftcomm "github.com/jialuohu/curlyraft/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

const (
	RPCTimeout = time.Millisecond * 100
)

func (rc *RaftCore) AppendEntries(ctx context.Context, request *raftcomm.AppendEntriesRequest) (*raftcomm.AppendEntriesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (rc *RaftCore) RequestVote(ctx context.Context, request *raftcomm.RequestVoteRequest) (*raftcomm.RequestVoteResponse, error) {
	rc.mu.Lock()

	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		log.Printf("[RPC Server] RequestVote: Error while getting current term: %v\n", err)
		return nil, err
	}
	candTerm := request.GetTerm()
	if curTerm > candTerm {
		rc.mu.Unlock()
		return &raftcomm.RequestVoteResponse{
			Term:        curTerm,
			VoteGranted: false,
		}, nil
	} else if curTerm < candTerm {
		rc.node.state = Follower
		if err := rc.node.setCurrentTerm(candTerm); err != nil {
			log.Printf("[RPC Server] RequestVote: Error while setting new term: %v\n", err)
			return nil, err
		}
	}

	votedFor, err := rc.node.getVotedFor()
	if err != nil {
		log.Printf("[RPC Server] RequestVote: Error while getting votedFor: %v\n", err)
		return nil, err
	}
	candId := request.GetCandidateId()
	candLastLogIndex := request.GetLastLogIndex()
	candLastLogTerm := request.GetLastLogTerm()
	if votedFor != VotedForNoOne &&
		votedFor != candId &&
		rc.node.lastLogIndex <= candLastLogIndex &&
		rc.node.lastLogTerm <= candLastLogTerm {

		rc.mu.Unlock()
		return &raftcomm.RequestVoteResponse{
			Term:        curTerm,
			VoteGranted: false,
		}, nil
	}

	if err := rc.node.setVotedFor(candId); err != nil {
		log.Printf("[RPC Server] RequestVote: Error while setting votedFor: %v\n", err)
		return nil, err
	}
	rc.mu.Unlock()
	return &raftcomm.RequestVoteResponse{
		Term:        curTerm,
		VoteGranted: true,
	}, nil
}

func (rc *RaftCore) InstallSnapShot(ctx context.Context, request *raftcomm.InstallSnapshotRequest) (*raftcomm.InstallSnapshotResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (rc *RaftCore) buildClient(netAddr string) (raftcomm.RaftCommunicationClient, *grpc.ClientConn, func(conn *grpc.ClientConn), error) {
	conn, err := grpc.NewClient(netAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("Failed to setup grpc channel:", err)
		return nil, nil, nil, err
	}
	closer := func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Println("Failed to close grpc channel:", err)
		}
	}
	client := raftcomm.NewRaftCommunicationClient(conn)
	return client, conn, closer, nil
}

func (rc *RaftCore) voteRequest(netAddr string, term uint32) {
	client, conn, closer, err := rc.buildClient(netAddr)
	if err != nil {
		log.Fatalf("[rpcClient] Failed to build grpc client: %v\n", err)
	}
	defer closer(conn)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
		res, err := client.RequestVote(ctx, &raftcomm.RequestVoteRequest{
			Term:         term,
			CandidateId:  rc.Info.id,
			LastLogIndex: rc.node.lastLogIndex,
			LastLogTerm:  rc.node.lastLogTerm,
		})
		cancel()

		if err != nil {
			log.Printf("[rpcClient] Failed to receive rpc response: %v\n", err)
			continue
		}

		rc.mu.Lock()
		if res.GetTerm() > term {
			if err := rc.node.setCurrentTerm(res.GetTerm()); err != nil {
				rc.mu.Unlock()
				log.Fatalf("[rpcClient] Failed to set new term: %v\n", err)
			}
		} else {
			if res.GetVoteGranted() {
				rc.receivedVotes++
				// check if the current candidate has enough votes out of all servers
				if rc.receivedVotes >= rc.quorumSize && rc.node.state == Candidate {
					rc.becomeLeader(term)
				}
			}
		}
		rc.mu.Unlock()
		return
	}
}

func (rc *RaftCore) heartbeatLoop() {
	for {
		log.Println("[Leader] heartbeating...")
		time.Sleep(time.Second)
	}
}
