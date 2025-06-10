package core

import (
	"context"
	"github.com/jialuohu/curlyraft/internal/clog"
	raftcomm "github.com/jialuohu/curlyraft/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

const (
	RPCTimeout = time.Millisecond * 100
	//RPCTimeout = time.Second * 10

	HeartbeatInterval = time.Millisecond * 50
	//HeartbeatInterval = time.Second * 1
)

func (rc *RaftCore) AppendEntries(ctx context.Context, request *raftcomm.AppendEntriesRequest) (*raftcomm.AppendEntriesResponse, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	log.Printf("%s The gRPC server tart handling AppendEntries RPC\n", clog.CGreenRc("AppendEntries"))
	lTerm := request.GetTerm()
	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		log.Printf("%s Error while getting current term: %v\n", clog.CRedRc("AppendEntries"), err)
		return nil, err
	}
	log.Printf("%s Start checking lTerm:%d, curTerm:%d\n",
		clog.CGreenRc("AppendEntries"), lTerm, curTerm)
	if lTerm < curTerm {
		log.Printf("%s lTerm < curTerm, leader is behind, will response false back\n", clog.CGreenRc("AppendEntries"))
		return &raftcomm.AppendEntriesResponse{
			Term:    curTerm,
			Success: false,
		}, nil
	} else if lTerm > curTerm {
		log.Printf("%s lTerm > curTerm, current node is behind\n", clog.CGreenRc("AppendEntries"))
		if err := rc.node.setCurrentTerm(lTerm); err != nil {
			log.Printf("%s Error while setting current term: %v\n", clog.CRedNode("AppendEntries"), err)
			return nil, err
		}
		rc.node.state = Follower
		rc.receivedVotes = 0

		log.Printf("%s tick heartbeatCh!\n", clog.CGreenRc("AppendEntries"))
		rc.node.tickHeartbeat()
	} else {
		log.Printf("%s tick heartbeatCh!\n", clog.CGreenRc("AppendEntries"))
		rc.node.tickHeartbeat()
	}

	// Consistency check
	log.Printf("%s Start checking consistency\n", clog.CBlueRc("AppendEntries"))
	lPrevIndex := request.GetPrevLogIndex()
	lPrevTerm := request.GetPrevLogTerm()

	log.Printf("%s lPrevIndex: %d, cur node lastLogIndex: %d\n",
		clog.CBlueRc("AppendEntries"), lPrevIndex, rc.node.lastLogIndex)
	if lPrevIndex > rc.node.lastLogIndex {
		log.Printf("%s lPrevIndex:%d > cur lastLogIndex:%d, cur node log missing and behind, response false\n",
			clog.CBlueRc("AppendEntries"), lPrevIndex, rc.node.lastLogIndex)
		return &raftcomm.AppendEntriesResponse{
			Term:    curTerm,
			Success: false,
		}, nil
	}
	logs, err := rc.node.getLog()
	log.Printf("%s Got Logs, its len:%d\n", clog.CGreenRc("AppendEntries"), len(logs))
	if err != nil {
		log.Printf("%s Error while getting logs: %v\n", clog.CRedRc("AppendEntries"), err)
		return nil, err
	}

	lEntries := request.GetEntries()
	if len(lEntries) != 0 {
		if logs[lPrevIndex].logTerm != lPrevTerm {
			log.Printf("%s lPrevTerm:%d != cur term in log on lPrevIndex:%d, cur node log missing, response false\n",
				clog.CGreenRc("AppendEntries"), lPrevTerm, logs[lPrevIndex].logTerm)
			return &raftcomm.AppendEntriesResponse{
				Term:    curTerm,
				Success: false,
			}, nil
		}
		log.Printf("%s Start deleting conflicts\n", clog.CGreenRc("AppendEntries"))
		if err := rc.node.deleteConflicts(lPrevIndex + 1); err != nil {
			log.Printf("%s Error while deleting conflicts: %v\n", clog.CRedRc("AppendEntries"), err)
			return nil, err
		}

		// Append new entries the leader sent
		log.Printf("%s Start append new entries\n", clog.CGreenRc("AppendEntries"))
		for _, e := range lEntries {
			entry := &logEntry{
				logTerm:  e.LogTerm,
				logIndex: e.LogIndex,
				command:  e.Command,
			}
			if err := rc.node.appendEntry(entry); err != nil {
				log.Printf("%s Error while append an entry %v: %v\n", clog.CRedRc("AppendEntries"), entry, err)
				return nil, err
			}
		}

		log.Printf("%s Setup cur node's lastLogIndex/Term\n", clog.CGreenRc("AppendEntries"))
		rc.node.lastLogIndex = lPrevIndex
		rc.node.lastLogTerm = lPrevTerm
	}

	// Advance commit index
	log.Printf("%s Start advancing commit index\n", clog.CGreenRc("AppendEntries"))
	lCommit := request.GetLeaderCommit()
	if lCommit > rc.node.commitIndex {
		log.Printf("%s leader's commitIdx: %d > cur node's commitIdx: %d\n",
			clog.CGreenRc("AppendEntries"), lCommit, rc.node.commitIndex)
		rc.node.commitIndex = min(lCommit, rc.node.lastLogIndex)
	}
	log.Printf("%s Applying new entries\n", clog.CGreenRc("AppendEntries"))
	rc.node.applyEntries()

	lId := request.GetLeaderId()
	log.Printf("%s Set new leaderId:%s in its state node\n", clog.CGreenRc("AppendEntries"), lId)
	rc.node.leaderId = rc.getPeerInfo(lId)
	log.Printf("%s Send true response back\n", clog.CGreenRc("AppendEntries"))
	return &raftcomm.AppendEntriesResponse{
		Term:    lCommit,
		Success: true,
	}, nil
}

func (rc *RaftCore) RequestVote(ctx context.Context, request *raftcomm.RequestVoteRequest) (*raftcomm.RequestVoteResponse, error) {
	log.Printf("%s The gRPC server start handling votes\n", clog.CGreenRc("RequestVote"))
	rc.mu.Lock()
	defer rc.mu.Unlock()

	log.Printf("%s The gRPC server get current term\n", clog.CGreenRc("RequestVote"))
	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		log.Printf("%s Error while getting current term: %v\n", clog.CRedRc("RequestVote"), err)
		return nil, err
	}
	candTerm := request.GetTerm()
	log.Printf("%s The gRPC server current term: %d, candidate term: %d\n", clog.CGreenRc("RequestVote"), curTerm, candTerm)
	if curTerm > candTerm {
		log.Printf("%s The current node term > candidate term, candidate is behind\n", clog.CGreenRc("RequestVote"))
		return &raftcomm.RequestVoteResponse{
			Term:        curTerm,
			VoteGranted: false,
		}, nil
	} else if curTerm < candTerm {
		log.Printf("%s The current node term < candidate term, current node is behind\n", clog.CGreenRc("RequestVote"))
		if rc.node.state == Leader {
			rc.node.leaderCancel()
		}
		rc.node.state = Follower
		if err := rc.node.setCurrentTerm(candTerm); err != nil {
			log.Printf("%s Error while setting new term: %v\n", clog.CRedRc("RequestVote"), err)
			return nil, err
		}
		if err := rc.node.setVotedFor(VotedForNoOne); err != nil {
			log.Printf("%s Error while setting votedFor as none: %v\n", clog.CRedRc("RequestVote"), err)
			return nil, err
		}
	}

	log.Printf("%s Start checking votedFor\n", clog.CGreenRc("RequestVote"))
	votedFor, err := rc.node.getVotedFor()
	if err != nil {
		log.Printf("%s Error while getting votedFor: %v\n", clog.CRedRc("RequestVote"), err)
		return nil, err
	}
	candId := request.GetCandidateId()
	candLastLogIndex := request.GetLastLogIndex()
	candLastLogTerm := request.GetLastLogTerm()
	log.Printf("%s current node votedFor: %s, candidate Id: %s, candidate LastLogIndex: %d, candidate LastLogTerm: %d\n",
		clog.CGreenRc("RequestVote"), votedFor, candId, candLastLogIndex, candLastLogTerm)
	if votedFor != VotedForNoOne &&
		votedFor != candId &&
		rc.node.lastLogIndex <= candLastLogIndex &&
		rc.node.lastLogTerm <= candLastLogTerm {

		log.Printf("%s Response false since 1).votedFor has another candidate 2).current ndoe last log (idx+term):%d+%d > candidate last log (idx+term)\n",
			clog.CGreenRc("RequestVote"), rc.node.lastLogIndex, rc.node.lastLogTerm)
		return &raftcomm.RequestVoteResponse{
			Term:        curTerm,
			VoteGranted: false,
		}, nil
	}

	log.Printf("%s set votedFor as candidate: %s\n", clog.CGreenRc("RequestVote"), candId)
	if err := rc.node.setVotedFor(candId); err != nil {
		log.Printf("%s Error while setting votedFor: %v\n", clog.CRedRc("RequestVote"), err)
		return nil, err
	}
	log.Printf("%s Response true back to candidate\n", clog.CGreenRc("RequestVote"))
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
	log.Printf("%s Start building gRPC client\n", clog.CGreenRc("buildClient"))
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
	log.Printf("%s Done building gRPC client\n", clog.CGreenRc("buildClient"))
	return client, conn, closer, nil
}

func (rc *RaftCore) voteRequest(netAddr string, term uint32) {
	log.Printf("%s Start voteRequest, netAddr: %s, term: %d\n",
		clog.CGreenRc("voteRequest"), netAddr, term)
	client, conn, closer, err := rc.buildClient(netAddr)
	if err != nil {
		log.Fatalf("[rpcClient] Failed to build grpc client: %v\n", err)
	}
	defer closer(conn)

	for {
		log.Printf("%s Sending RequestVote RPC request to %s, lastLogIndex: %d, lastLogTerm: %d\n",
			clog.CGreenRc("voteRequest"), netAddr, rc.node.lastLogIndex, rc.node.lastLogTerm)
		ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
		res, err := client.RequestVote(ctx, &raftcomm.RequestVoteRequest{
			Term:         term,
			CandidateId:  rc.Info.id,
			LastLogIndex: rc.node.lastLogIndex,
			LastLogTerm:  rc.node.lastLogTerm,
		})
		cancel()

		if err != nil {
			log.Printf("%s Failed to receive rpc response: %v\n", clog.CRedRc("voteRequest"), err)
			time.Sleep(RPCTimeout)
			continue
		}

		rc.mu.Lock()
		log.Printf("%s Checking response from other node: %s\n",
			clog.CGreenRc("voteRequest"), netAddr)
		if res.GetTerm() > term {
			log.Printf("%s peer node term: %d > cur term: %d\n",
				clog.CGreenRc("voteRequest"), res.GetTerm(), term)
			if err := rc.node.setCurrentTerm(res.GetTerm()); err != nil {
				rc.mu.Unlock()
				log.Fatalf("%s Failed to set new term: %v\n", clog.CRedRc("voteRequest"), err)
			}
			log.Printf("%s Updated cur term as peer term\n",
				clog.CGreenRc("voteRequest"))
		} else {
			log.Printf("%s peer node term: %d <= cur term: %d\n",
				clog.CGreenRc("voteRequest"), res.GetTerm(), term)
			if res.GetVoteGranted() {
				log.Printf("%s Got one vote from peer!\n", clog.CGreenRc("voteRequest"))
				rc.receivedVotes++
				// check if the current candidate has enough votes out of all servers
				if rc.receivedVotes >= rc.quorumSize && rc.node.state == Candidate {
					log.Printf("%s Have enough votes: %d! Gonna become leader!\n",
						clog.CBlueRc("voteRequest"), rc.receivedVotes)
					rc.becomeLeader(term)
				}
			} else {
				log.Printf("%s Didn't get vote from peer\n", clog.CGreenRc("voteRequest"))
			}
		}
		log.Printf("%s Done with voteRequest\n", clog.CGreenRc("voteRequest"))
		rc.mu.Unlock()
		return
	}
}

func (rc *RaftCore) sendHeartbeat(ctx context.Context, client raftcomm.RaftCommunicationClient, peer nodeInfo, term uint32) {

	rc.mu.Lock()
	// since heartbeat just send nothing
	//prevLogIndex := rc.node.nextIndex[peer.id] - 1
	//prevLogTerm, ok := rc.node.getLogTerm(prevLogIndex)

	prevLogIndex := rc.node.lastLogIndex
	prevLogTerm := rc.node.lastLogTerm

	//if !ok {
	//	log.Fatalf("%s Failed to get log term based on idx:%d.\n",
	//		clog.CRedRc("sendHeartbeat"), prevLogIndex)
	//}

	commitIndex := rc.node.commitIndex
	log.Printf("%s prevLogIndex: %d, prevLogTerm: %d, commitIdx: %d\n",
		clog.CGreenRc("sendHeartbeat"), prevLogIndex, prevLogTerm, commitIndex)
	rc.mu.Unlock()

	rpcCtx, cancel := context.WithTimeout(ctx, RPCTimeout)
	defer cancel()

	log.Printf("%s Send AppendEntries RPC request (heartbeat) to %s\n",
		clog.CGreenRc("sendHeartbeat"), peer.id)
	res, err := client.AppendEntries(rpcCtx, &raftcomm.AppendEntriesRequest{
		Term:         term,
		LeaderId:     rc.Info.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      nil,
		LeaderCommit: commitIndex,
	})
	if err != nil {
		log.Printf("%s Heartbeat request to %s failed: %v", clog.CRedRc("sendHeartbeat"), peer.netAddr, err)
		return
	}

	rc.mu.Lock()
	defer rc.mu.Unlock()

	log.Printf("%s Check term\n", clog.CGreenRc("sendHeartbeat"))
	nodeTerm := res.GetTerm()
	if nodeTerm > term {
		log.Printf("%s saw higher term %d > %d, stepping down", clog.CGreenRc("sendHeartbeat"), nodeTerm, term)
		rc.node.leaderCancel()
		rc.node.state = Follower
		rc.node.matchIndex = nil
		rc.node.nextIndex = nil
		if err := rc.node.setCurrentTerm(nodeTerm); err != nil {
			log.Printf("%s Failed to set current term to peer's\n", clog.CGreenRc("sendHeartbeat"))
		}
		return
	}

	isSuccess := res.GetSuccess()
	if isSuccess {
		// follower accepted
		log.Printf("%s peer %s accepted heartbeat\n", clog.CGreenRc("sendHeartbeat"), peer.id)
		//rc.node.nextIndex[peer.id] = prevLogIndex + 1
		//rc.node.matchIndex[peer.id] = prevLogIndex
	} else {
		// follower rejected
		log.Printf("%s peer %s rejected heartbeat\n", clog.CGreenRc("sendHeartbeat"), peer.id)
		//if rc.node.nextIndex[peer.id] > 1 {
		//	rc.node.nextIndex[peer.id]--
		//}
	}
	//log.Printf("%s nextIndex, matchIndex for peer %s: %d, %d\n",
	//	clog.CGreenRc("sendHeartbeat"), peer.id, rc.node.nextIndex[peer.id], rc.node.matchIndex[peer.id])
}

func (rc *RaftCore) heartbeatLoop(ctx context.Context, term uint32) {
	type clientConnInfo struct {
		conn   *grpc.ClientConn
		closer func(conn *grpc.ClientConn)
	}

	peerConns := make([]clientConnInfo, 0, len(rc.Peers))
	for _, p := range rc.Peers {
		log.Printf("%s Start sending heartbeat to %s/%s\n", clog.CGreenRc("heartbeatLoop"), p.id, p.netAddr)
		client, conn, closer, err := rc.buildClient(p.netAddr)
		if err != nil {
			log.Fatalf("%s Failed to build grpc client: %v\n", clog.CRedRc("heartbeatLoop"), err)
		}
		peerConns = append(peerConns, clientConnInfo{
			conn:   conn,
			closer: closer,
		})
		go rc.sendHeartbeat(ctx, client, p, term)
	}
	defer func(peerConns []clientConnInfo) {
		for _, c := range peerConns {
			c.closer(c.conn)
		}
	}(peerConns)

	log.Printf("%s Start a new ticker for heartbeat\n", clog.CGreenRc("heartbeatLoop"))
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// leaderCtx was canceled
			log.Printf("%s heartbeatLoop exiting\n", clog.CGreenRc("heartbeatLoop"))
			return

		case <-ticker.C:
			// send actual AppendEntries heartbeats here
			log.Printf("%s heartbeating…\n", clog.CGreenRc("heartbeatLoop"))
			for i, p := range rc.Peers {
				client := raftcomm.NewRaftCommunicationClient(peerConns[i].conn)
				go rc.sendHeartbeat(ctx, client, p, term)
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
