package core

import (
	"context"
	"github.com/jialuohu/curlyraft/internal/clog"
	"log"
	"math/rand"
	"time"
)

const (
	TimeoutLowerBound = 150 * time.Millisecond
	//TimeoutLowerBound = 5 * time.Second

	TimeoutRange = 150 * time.Millisecond
	//TimeoutRange = 5 * time.Second
)

func (rc *RaftCore) roleLoop() error {
	for {
		log.Printf("%s Start choosing random timeout value\n", clog.CGreenRc("roleLoop"))
		jitter := time.Duration(rand.Int63n(int64(TimeoutRange)))
		timeout := TimeoutLowerBound + jitter

		if rc.node.timer == nil {
			log.Printf("%s Create random timer\n", clog.CGreenRc("roleLoop"))
			rc.node.timer = time.NewTimer(timeout)
		} else {
			log.Printf("%s Restart random timer\n", clog.CGreenRc("roleLoop"))
			rc.node.timer.Stop()
			rc.node.timer.Reset(timeout)
		}

		select {
		case <-rc.node.timer.C:
			rc.mu.Lock()
			notLeader := rc.node.state == Follower || rc.node.state == Candidate
			rc.mu.Unlock()

			if notLeader {
				log.Printf("%s Timeout! Start election\n", clog.CBlueRc("roleLoop"))
				if err := rc.startElection(); err != nil {
					log.Printf("%s Error during startElection: %v\n", clog.CRedRc("roleLoop"), err)
					return err
				}
				log.Printf("%s Finished election\n", clog.CGreenRc("roleLoop"))
			} else {
				log.Printf("%s Timeout! But we are Leader! So skipping!\n", clog.CBlueRc("roleLoop"))
			}
		case <-rc.node.heartbeatCh:
			log.Printf("%s Received Heartbeat from a leader\n", clog.CBlueRc("roleLoop"))
			continue
		case <-rc.node.stopCh:
			log.Printf("%s Received Stop signal\n", clog.CBlueRc("roleLoop"))
			if rc.node.timer != nil {
				rc.node.timer.Stop()
			}
			return nil
		}
	}
}

func (n *node) tickHeartbeat() {
	log.Printf("%s Tick heartbeat!\n", clog.CGreenNode("tickHeartbeat"))
	select {
	case n.heartbeatCh <- struct{}{}:
		log.Printf("%s Sent into heartbeat channel\n", clog.CGreenNode("tickHeartbeat"))
	default:
		log.Printf("%s Heartbeat channel is full\n", clog.CGreenNode("tickHeartbeat"))
		// so that it will not block RPC handler
	}
}

func (rc *RaftCore) startElection() error {
	log.Printf("%s Start election\n", clog.CGreenRc("startElection"))
	rc.mu.Lock()
	log.Printf("%s Get current node term\n", clog.CGreenRc("startElection"))
	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		return err
	}
	newTerm := curTerm + 1
	if err := rc.node.setCurrentTerm(newTerm); err != nil {
		return err
	}
	log.Printf("%s Set current node term: %d, newTerm: %d\n", clog.CGreenRc("startElection"), curTerm, newTerm)

	if rc.node.state == Follower {
		log.Printf("%s Current node state is Follower will become Candidate\n", clog.CGreenRc("startElection"))
		rc.node.state = Candidate
		rc.receivedVotes++
		log.Printf("%s Start sending voteRequest to its peers\n", clog.CGreenRc("startElection"))
		for _, peer := range rc.Peers {
			rc.node.leaderBecomeCtx, rc.node.leaderBecomeCancel = context.WithCancel(context.Background())
			go rc.voteRequest(rc.node.leaderBecomeCtx, peer.netAddr, newTerm)
		}
		log.Printf("%s Done with sending\n", clog.CGreenRc("startElection"))
	} else if rc.node.state == Candidate {
		log.Printf("%s Current node state is Candidate\n", clog.CGreenRc("startElection"))
		rc.receivedVotes = 1 // empty previous term votes + vote itself
		log.Printf("%s Start sending voteRequest to its peers\n", clog.CGreenRc("startElection"))
		for _, peer := range rc.Peers {
			rc.node.leaderBecomeCtx, rc.node.leaderBecomeCancel = context.WithCancel(context.Background())
			go rc.voteRequest(rc.node.leaderBecomeCtx, peer.netAddr, newTerm)
		}
	}

	log.Printf("%s Node votedFor itself!!!\n", clog.CBlueRc("startElection"))
	if err := rc.node.setVotedFor(rc.Info.id); err != nil {
		log.Printf("%s Failed to set votedFor\n", clog.CRedRc("startElection"))
	}

	rc.mu.Unlock()
	return nil
}

func (rc *RaftCore) becomeLeader(term uint32) {
	// caller must hold rc.mu
	log.Printf("%s Node %s becomes Leader in term %d\n", clog.CGreenRc("becomeLeader"), rc.Info.id, term)

	rc.node.state = Leader
	size := len(rc.Peers) + 1

	rc.node.nextIndex = make(map[string]uint32, size)
	rc.node.matchIndex = make(map[string]uint32, size)
	for _, p := range rc.Peers {
		rc.node.nextIndex[p.id] = rc.node.lastLogIndex + 1
		rc.node.matchIndex[p.id] = 0
	}

	rc.node.leaderCtx, rc.node.leaderCancel = context.WithCancel(context.Background())
	// Once the node becomes leader, will call cancel.
	rc.node.leaderBecomeCancel()

	log.Printf("%s Start heartbeatLoop\n", clog.CGreenRc("becomeLeader"))
	go rc.heartbeatLoop(rc.node.leaderCtx, term)
	log.Printf("%s Start leader start serving\n", clog.CGreenRc("becomeLeader"))
	go func() {
		if err := rc.leaderStartServing(rc.node.leaderCtx); err != nil {
			log.Fatalf("%s Leader serving failed\n", clog.CRedRc("becomeLeader"))
		}
	}()
}
