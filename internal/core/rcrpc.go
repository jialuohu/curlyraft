package core

import (
	"context"
	"fmt"
	"github.com/jialuohu/curlyraft/internal/clog"
	"github.com/jialuohu/curlyraft/internal/proto/gateway"
	"github.com/jialuohu/curlyraft/internal/proto/raftcomm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math"
	"net"
	"os"
	"time"
)

const (
	RPCTimeout = time.Millisecond * 50

	HeartbeatInterval = time.Millisecond * 80
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
		rc.node.lastLogIndex = lPrevIndex + uint32(len(lEntries))
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

	// Setup leader id for current follower node
	lId := request.GetLeaderId()
	log.Printf("%s Set leaderId:%s in its state node\n", clog.CGreenRc("AppendEntries"), lId)
	getPeerInfo := func(nodeId string) *nodeInfo {
		for _, p := range rc.Peers {
			if p.id == nodeId {
				return &p
			}
		}
		return nil
	}
	rc.node.leaderId = getPeerInfo(lId)
	if rc.node.leaderId == nil {
		return nil, fmt.Errorf("%s The leaderId is nil!!!\n", clog.CRedRc("AppendEntries"))
	}

	if err := rc.node.setVotedFor(request.GetLeaderId()); err != nil {
		log.Printf("%s Error while setting votedFor as leader id: %v\n", clog.CRedNode("AppendEntries"), err)
		return nil, err
	}

	log.Printf("%s Send true response back\n", clog.CGreenRc("AppendEntries"))
	return &raftcomm.AppendEntriesResponse{
		Term:    lTerm,
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

func (rc *RaftCore) Check(ctx context.Context, request *raftcomm.HealthCheckRequest) (*raftcomm.HealthCheckResponse, error) {
	service := request.GetService()
	if service == "run" {
		return &raftcomm.HealthCheckResponse{Status: raftcomm.HealthCheckResponse_SERVING}, nil
	} else {
		return &raftcomm.HealthCheckResponse{Status: raftcomm.HealthCheckResponse_NOT_SERVING}, nil
	}
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

func (rc *RaftCore) voteRequest(ctx context.Context, netAddr string) {
	log.Printf("%s Start voteRequest, netAddr: %s\n",
		clog.CGreenRc("voteRequest"), netAddr)
	ci, err := rc.getHealthyConn(netAddr)
	if err != nil {
		log.Printf("%s Failed to build grpc client: %v\n", clog.CRedRc("voteRequest"), err)
	}

	sendRequest := func(ci *clientConnInfo) *raftcomm.RequestVoteResponse {
		log.Printf("%s Sending RequestVote RPC request to %s, lastLogIndex: %d, lastLogTerm: %d\n",
			clog.CGreenRc("voteRequest"), netAddr, rc.node.lastLogIndex, rc.node.lastLogTerm)
		if ci == nil {
			return nil
		}
		rc.mu.Lock()
		curTerm, err := rc.node.getCurrentTerm()
		if err != nil {
			log.Fatalf("%s Failed to get current term: %v\n", clog.CRedRc("voteRequest"), err)
		}
		lastLogIndex := rc.node.lastLogIndex
		lastLogTerm := rc.node.lastLogTerm
		rc.mu.Unlock()
		res, err := ci.client.RequestVote(ctx, &raftcomm.RequestVoteRequest{
			Term:         curTerm,
			CandidateId:  rc.Info.id,
			LastLogIndex: lastLogIndex,
			LastLogTerm:  lastLogTerm,
		})
		if err != nil {
			log.Printf("%s Failed to receive rpc response: %v\n", clog.CRedRc("voteRequest"), err)
			return nil
		}
		return res
	}

	checkResponse := func(res *raftcomm.RequestVoteResponse) bool {
		if res == nil {
			return false
		}

		rc.mu.Lock()
		curTerm, err := rc.node.getCurrentTerm()
		if err != nil {
			log.Fatalf("%s Failed to get current term: %v\n", clog.CRedRc("voteRequest"), err)
		}
		log.Printf("%s Checking response from other node: %s\n",
			clog.CGreenRc("voteRequest"), netAddr)
		if res.GetTerm() > curTerm {
			log.Printf("%s peer node term: %d > cur term: %d\n",
				clog.CGreenRc("voteRequest"), res.GetTerm(), curTerm)
			if err := rc.node.setCurrentTerm(res.GetTerm()); err != nil {
				rc.mu.Unlock()
				log.Fatalf("%s Failed to set new term: %v\n", clog.CRedRc("voteRequest"), err)
			}
			log.Printf("%s Updated cur term as peer term\n",
				clog.CGreenRc("voteRequest"))
		} else {
			log.Printf("%s peer node term: %d <= cur term: %d\n",
				clog.CGreenRc("voteRequest"), res.GetTerm(), curTerm)
			if res.GetVoteGranted() {
				log.Printf("%s Got one vote from peer!\n", clog.CGreenRc("voteRequest"))
				rc.receivedVotes++
				// check if the current candidate has enough votes out of all servers
				if rc.receivedVotes >= rc.quorumSize && rc.node.state == Candidate {
					log.Printf("%s Have enough votes: %d! Gonna become leader!\n",
						clog.CBlueRc("voteRequest"), rc.receivedVotes)
					rc.becomeLeader(curTerm)
				}
			} else {
				log.Printf("%s Didn't get vote from peer\n", clog.CGreenRc("voteRequest"))
			}
		}
		log.Printf("%s Done with voteRequest\n", clog.CGreenRc("voteRequest"))
		rc.mu.Unlock()
		return true
	}

	if checkResponse(sendRequest(ci)) {
		return
	}
	ticker := time.NewTicker(RPCTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ci, err = rc.getHealthyConn(netAddr)
			if err != nil {
				log.Printf("%s Failed to build grpc client: %v\n", clog.CRedRc("voteRequest"), err)
				continue
			}
			if checkResponse(sendRequest(ci)) {
				return
			}
		case <-ctx.Done():
			log.Printf("%s Already leader, no need to keep sending voteRequest\n", clog.CBlueRc("voteRequest"))
			return
		}
	}
}

func (rc *RaftCore) sendHeartbeat(ctx context.Context, client raftcomm.RaftCommunicationClient, peer nodeInfo, term uint32) {
	rc.mu.Lock()
	prevLogIndex := rc.node.nextIndex[peer.id] - 1
	prevLogTerm, found := rc.node.getLogTerm(prevLogIndex)
	if !found {
		log.Fatalf("%s Failed to get term based on prevLogIndex:%d\n",
			clog.CRedRc("sendHeartbeat"), prevLogIndex)
	}
	commitIndex := rc.node.commitIndex
	log.Printf("%s prevLogIndex: %d, prevLogTerm: %d, commitIdx: %d\n",
		clog.CGreenRc("sendHeartbeat"), prevLogIndex, prevLogTerm, commitIndex)
	rc.mu.Unlock()

	//rpcCtx, cancel := context.WithTimeout(ctx, RPCTimeout)
	//defer cancel()

	log.Printf("%s Send AppendEntries RPC request (heartbeat) to %s\n",
		clog.CGreenRc("sendHeartbeat"), peer.id)
	res, err := client.AppendEntries(ctx, &raftcomm.AppendEntriesRequest{
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
		if err := rc.leaderStepdown(nodeTerm); err != nil {
			log.Printf("%s Failed to stepdown leader: %v\n", clog.CRedRc("sendHeartbeat"), err)
		}
		return
	}

	isSuccess := res.GetSuccess()
	if isSuccess {
		// follower accepted
		rc.node.matchIndex[peer.id] = prevLogIndex
		log.Printf("%s peer %s accepted heartbeat, matchIndex: %d\n",
			clog.CGreenRc("sendHeartbeat"), peer.id, rc.node.matchIndex[peer.id])
	} else {
		// follower rejected
		rc.node.nextIndex[peer.id]--
		log.Printf("%s peer %s rejected heartbeat, nextIndex--:%d\n",
			clog.CGreenRc("sendHeartbeat"), peer.id, rc.node.nextIndex[peer.id])
	}
}

func (rc *RaftCore) getHealthyConn(netAddr string) (*clientConnInfo, error) {
	rc.mu.Lock()
	ci := rc.peerConns[netAddr]
	rc.mu.Unlock()

	if ci == nil {
		client, conn, closer, err := rc.buildClient(netAddr)
		if err != nil {
			return nil, fmt.Errorf("cannot build up conn with %s", netAddr)
		}
		ci = &clientConnInfo{client: client, conn: conn, closer: closer}
		rc.mu.Lock()
		rc.peerConns[netAddr] = ci
		rc.mu.Unlock()
		return ci, nil
	}

	state := ci.conn.GetState()
	if state != connectivity.Ready && state != connectivity.Idle {
		// tear down and rebuild
		ci.closer(ci.conn)
		client, conn, closer, err := rc.buildClient(netAddr)
		if err != nil {
			return nil, err
		}
		ci = &clientConnInfo{client: client, conn: conn, closer: closer}
		rc.mu.Lock()
		rc.peerConns[netAddr] = ci
		rc.mu.Unlock()
	}
	return ci, nil
}

func (rc *RaftCore) heartbeatLoop(ctx context.Context, term uint32) {
	for _, p := range rc.Peers {
		log.Printf("%s Start sending heartbeat to %s/%s\n", clog.CGreenRc("heartbeatLoop"), p.id, p.netAddr)
		ci, err := rc.getHealthyConn(p.netAddr)
		if err != nil {
			log.Printf("%s Error while performed getHealthyConn, %v\n", clog.CRedRc("heartbeatLoop"), err)
			continue
		}
		go rc.sendHeartbeat(ctx, ci.client, p, term)
	}

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
			for _, p := range rc.Peers {
				ci, err := rc.getHealthyConn(p.netAddr)
				if err != nil {
					log.Printf("%s Error while performed getHealthyConn, %v\n", clog.CRedRc("heartbeatLoop"), err)
					continue
				}
				go rc.sendHeartbeat(ctx, ci.client, p, term)
			}
		}
	}
}

func (rc *RaftCore) leaderStartServing(ctx context.Context) error {
	socketPath := "/tmp/raft.sock"
	if err := os.RemoveAll(socketPath); err != nil {
		log.Printf("[leaderServing] warning removing stale socket: %v", err)
	}

	log.Printf("%s Leader start serving \n", clog.CGreenRc("leaderServing"))
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Printf("%s Error while creating listener on /tmp/raft.sock:%v\n", clog.CRedRc("leaderStartServing"), err)
		return err
	}
	grpcServer := grpc.NewServer()
	rc.mu.Lock()
	rc.rg = NewRaftGateway(rc, grpcServer)
	rc.mu.Unlock()

	gateway.RegisterRaftGatewayServer(grpcServer, rc.rg)
	gateway.RegisterHealthCheckServer(grpcServer, rc.rg)
	go func(ctx context.Context) {
		<-ctx.Done()
		log.Printf("%s Leader gateway gracefully stops\n", clog.CGreenRc("leaderServing"))
		rc.rg.stop()
	}(ctx)
	return grpcServer.Serve(lis)
}

func (rc *RaftCore) entriesSend(ctx context.Context, ci *clientConnInfo, p nodeInfo) (uint32, bool, uint32) {
	log.Printf("%s Start sending entries\n", clog.CBlueRc("entriesSend"))
	rc.mu.Lock()
	leaderId := rc.Info.id
	prevLogIndex := rc.node.nextIndex[p.id] - 1
	leaderCommit := rc.node.commitIndex
	log.Printf("%s leaderId:%s, nextIdx[peer]:%d, prevLogIndex:%d, leaderCommit:%d\n",
		clog.CBlueRc("entriesSend"), leaderId, rc.node.nextIndex[p.id], prevLogIndex, leaderCommit)
	rc.mu.Unlock()
	prevLogTerm, found := rc.node.getLogTerm(prevLogIndex)
	if !found {
		log.Fatalf("%s Failed to get term based on index:%d\n",
			clog.CRedRc("entriesSend"), prevLogIndex)
	}
	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		log.Fatalf("%s Failed to get current term\n", clog.CRedRc("entriesSend"))
	}
	logEntries, err := rc.node.getLog()
	if err != nil {
		log.Fatalf("%s Failed to get log list\n", clog.CRedRc("entriesSend"))
	}

	entries := make([]*raftcomm.LogEntry, 0, len(logEntries))
	for _, e := range logEntries[prevLogIndex+1:] {
		entries = append(entries, &raftcomm.LogEntry{
			Command:  e.command,
			LogTerm:  e.logTerm,
			LogIndex: e.logIndex,
		})
	}
	log.Printf("%s len(entries):%d, len(logEntries):%d, entries: %v\n",
		clog.CBlueRc("entriesSend"), len(entries), len(logEntries), entries)

	log.Printf("%s Prepared done, now start sending RPC request\n", clog.CBlueRc("entriesSend"))
	res, err := ci.client.AppendEntries(ctx, &raftcomm.AppendEntriesRequest{
		Term:         curTerm,
		LeaderId:     leaderId,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		LeaderCommit: leaderCommit,
		Entries:      entries,
	})
	if err != nil {
		log.Printf("%s Failed to get response from follower %s/%s: %v\n",
			clog.CRedRc("entriesSend"), p.id, p.netAddr, err)
	}
	if res != nil {
		log.Printf("%s Got RPC response successfully\n", clog.CBlueRc("entriesSend"))
		return res.GetTerm(), res.GetSuccess(), uint32(len(entries))
	}
	return 0, false, 0
}

func (rc *RaftCore) bCastEntriesToPeers(ctx context.Context) {
	responseProcess := func(p nodeInfo, fTerm uint32, success bool, numEntries uint32) bool {
		if success {
			log.Printf("%s Appended entries into follower successfully, fTerm:%d\n",
				clog.CGreenRc("bCastEntriesToPeers"), fTerm)
			rc.mu.Lock()
			log.Printf("%s Append successfully, updating matchIndex:%d=nextIndex-1...\n",
				clog.CBlueRc("bCastEntriesToPeers"), rc.node.nextIndex[p.id]-1)
			rc.node.nextIndex[p.id] += numEntries
			rc.node.matchIndex[p.id] += numEntries
			log.Printf("%s Now matchIndex:%d, nextIndex:%d\n",
				clog.CBlueRc("bCastEntriesToPeers"), rc.node.matchIndex[p.id], rc.node.nextIndex[p.id])
			rc.mu.Unlock()
			return true
		} else {
			log.Printf("%s Appending failed\n", clog.CGreenRc("entriesSend"))
			curTerm, err := rc.node.getCurrentTerm()
			if err != nil {
				log.Printf("%s Failed to get current term\n", clog.CRedRc("bCastEntriesToPeers"))
				return false
			}
			if curTerm < fTerm {
				log.Printf("%s saw higher term %d > %d, stepping down",
					clog.CGreenRc("bCastEntriesToPeers"), fTerm, curTerm)
				if err = rc.leaderStepdown(fTerm); err != nil {
					log.Printf("%s Failed to make leader stepdown\n", clog.CRedRc("bCastEntriesToPeers"))
				}
			} else {
				rc.mu.Lock()
				log.Printf("%s Append failed, nextIdx--:%d\n",
					clog.CGreenRc("bCastEntriesToPeers"), rc.node.nextIndex[p.id]-1)
				rc.node.nextIndex[p.id]--
				rc.mu.Unlock()
			}
			return false
		}
	}

	rpcRetransmit := func(ci *clientConnInfo, p nodeInfo) {
		fTerm, success, numEntries := rc.entriesSend(ctx, ci, p)
		if responseProcess(p, fTerm, success, numEntries) {
			return
		}
		t := time.NewTicker(RPCTimeout)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Printf("%s Timeout! Resend RPC request to follower %s/%s\n",
					clog.CBlueRc("bCastEntriesToPeers"), p.id, p.netAddr)
				fTerm, success, numEntries = rc.entriesSend(ctx, ci, p)
				if responseProcess(p, fTerm, success, numEntries) {
					return
				}
			case <-ctx.Done():
				log.Printf("%s Leader stepping down, no need to send AppendEntries RPC request\n",
					clog.CBlueRc("bCastEntriesToPeers"))
				return
			}
		}
	}

	for _, p := range rc.Peers {
		log.Printf("%s Start sending entries to %s/%s\n",
			clog.CGreenRc("bCastEntriesToPeers"), p.id, p.netAddr)
		ci, err := rc.getHealthyConn(p.netAddr)
		if err != nil {
			log.Printf("%s Error while performed getHealthyConn, %v\n", clog.CRedRc("bCastEntriesToPeers"), err)
			continue
		}

		go rpcRetransmit(ci, p)
	}
}

func (rc *RaftCore) updateCommitIndex() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	minIdx := uint32(math.MaxUint32)
	for _, v := range rc.node.matchIndex {
		if v < minIdx {
			minIdx = v
		}
	}
	rc.node.commitIndex = minIdx
}

func (rc *RaftCore) propose(cmd []byte) error {
	rc.mu.Lock()
	log.Printf("%s Start propose cmd into log storage\n", clog.CBlueRc("propose"))
	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		log.Printf("%s Failed to get current term\n", clog.CRedRc("propose"))
		return err
	}
	if rc.node.lastLogTerm < curTerm {
		log.Printf("%s Also update lastLogTerm\n", clog.CGreenRc("propose"))
		rc.node.lastLogTerm = curTerm
	}
	le := &logEntry{
		logTerm:  rc.node.lastLogTerm,
		logIndex: rc.node.lastLogIndex + 1,
		command:  cmd,
	}
	if err := rc.node.appendEntry(le); err != nil {
		log.Printf("%s Failed to append entry into log\n", clog.CRedRc("propose"))
	}
	rc.node.lastLogIndex++
	log.Printf("%s Done with propose. Now lastLogTerm:%d,lastLogIdx:%d\n",
		clog.CGreenRc("propose"), rc.node.lastLogTerm, rc.node.lastLogIndex)
	rc.mu.Unlock()

	rc.bCastEntriesToPeers(rc.node.leaderCtx)
	rc.updateCommitIndex()
	rc.node.applyEntries()
	return nil
}
