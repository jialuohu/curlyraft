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
	RPCTimeout        = time.Millisecond * 100
	HeartbeatInterval = time.Millisecond * 50
)

func (rc *RaftCore) AppendEntries(ctx context.Context, request *raftcomm.AppendEntriesRequest) (*raftcomm.AppendEntriesResponse, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	lTerm := request.GetTerm()
	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		log.Printf("[RPC Server] AppendEntries: Error while getting current term: %v\n", err)
		return nil, err
	}
	if lTerm < curTerm {
		return &raftcomm.AppendEntriesResponse{
			Term:    curTerm,
			Success: false,
		}, nil
	} else if lTerm > curTerm {
		if err := rc.node.setCurrentTerm(lTerm); err != nil {
			log.Printf("[RPC Server] AppendEntries: Error while setting new term: %v\n", err)
			return nil, err
		}
		rc.node.state = Follower
		rc.receivedVotes = 0
		rc.node.tickHeartbeat()
	}

	// Consistency check
	lPrevIndex := request.GetPrevLogIndex()
	lPrevTerm := request.GetPrevLogTerm()

	if lPrevIndex > rc.node.lastLogIndex {
		return &raftcomm.AppendEntriesResponse{
			Term:    curTerm,
			Success: false,
		}, nil
	}
	logs, err := rc.node.getLog()
	if err != nil {
		log.Printf("[RPC Server] AppendEntries: Error while getting logs: %v\n", err)
		return nil, err
	}
	if logs[lPrevIndex].logTerm != lPrevTerm {
		return &raftcomm.AppendEntriesResponse{
			Term:    curTerm,
			Success: false,
		}, nil
	}
	if err := rc.node.deleteConflicts(lPrevIndex + 1); err != nil {
		log.Printf("[RPC Server] AppendEntries: Error while deleting conflicts: %v\n", err)
		return nil, err
	}
	rc.node.lastLogIndex = lPrevIndex
	rc.node.lastLogTerm = lPrevTerm

	// Append new entries the leader sent
	lEntries := request.GetEntries()
	for _, e := range lEntries {
		entry := &logEntry{
			logTerm:  e.LogTerm,
			logIndex: e.LogIndex,
			command:  e.Command,
		}
		if err := rc.node.appendEntry(entry); err != nil {
			return nil, err
		}
	}

	// Advance commit index
	lCommit := request.GetLeaderCommit()
	if lCommit > rc.node.commitIndex {
		rc.node.commitIndex = min(lCommit, rc.node.lastLogIndex)
	}
	rc.node.applyEntries()

	lId := request.GetLeaderId()
	rc.node.leaderId = rc.getPeerInfo(lId)
	return &raftcomm.AppendEntriesResponse{
		Term:    lCommit,
		Success: true,
	}, nil
}

func (rc *RaftCore) RequestVote(ctx context.Context, request *raftcomm.RequestVoteRequest) (*raftcomm.RequestVoteResponse, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		log.Printf("[RPC Server] RequestVote: Error while getting current term: %v\n", err)
		return nil, err
	}
	candTerm := request.GetTerm()
	if curTerm > candTerm {
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

		return &raftcomm.RequestVoteResponse{
			Term:        curTerm,
			VoteGranted: false,
		}, nil
	}

	if err := rc.node.setVotedFor(candId); err != nil {
		log.Printf("[RPC Server] RequestVote: Error while setting votedFor: %v\n", err)
		return nil, err
	}
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

func (rc *RaftCore) sendHeartbeat(ctx context.Context, peer nodeInfo, term uint32) {
	client, conn, closer, err := rc.buildClient(peer.netAddr)
	if err != nil {
		log.Fatalf("[rpcClient] Failed to build grpc client: %v\n", err)
	}
	defer closer(conn)

	rc.mu.Lock()
	prevLogIndex := rc.node.nextIndex[peer.id]
	prevLogTerm, ok := rc.node.getLogTerm(prevLogIndex)
	if !ok {
		log.Fatalf("[sendHeartbeat] Failed to get log term based on idx.\n")
	}
	commitIndex := rc.node.commitIndex
	rc.mu.Unlock()

	rpcCtx, cancel := context.WithTimeout(ctx, RPCTimeout)
	defer cancel()

	res, err := client.AppendEntries(rpcCtx, &raftcomm.AppendEntriesRequest{
		Term:         term,
		LeaderId:     rc.Info.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      nil,
		LeaderCommit: commitIndex,
	})
	if err != nil {
		log.Printf("[sendHeartbeat] AppendEntries to %s failed: %v", peer.netAddr, err)
		return
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	nodeTerm := res.GetTerm()
	if nodeTerm > term {
		log.Printf("[sendHeartbeat] saw higher term %d > %d, stepping down", nodeTerm, term)
		rc.node.state = Follower
		rc.node.leaderCancel()
		return
	}

	isSuccess := res.GetSuccess()
	if isSuccess {
		// follower accepted
		rc.node.nextIndex[peer.id] = prevLogIndex + 1
		rc.node.matchIndex[peer.id] = prevLogIndex
	} else {
		// follower rejected
		if rc.node.nextIndex[peer.id] > 1 {
			rc.node.nextIndex[peer.id]--
		}
	}
}

func (rc *RaftCore) heartbeatLoop(ctx context.Context, term uint32) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// leaderCtx was canceled
			log.Println("[Leader] heartbeatLoop exiting")
			return

		case <-ticker.C:
			// send actual AppendEntries heartbeats here
			log.Println("[Leader] heartbeating…")
			for _, p := range rc.Peers {
				go rc.sendHeartbeat(ctx, p, term)
			}
		}
	}
}

func (rc *RaftCore) getPeerInfo(nodeId string) *nodeInfo {
	for _, p := range rc.Peers {
		if p.id == nodeId {
			return &p
		}
	}
	return nil
}
