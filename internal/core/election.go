package core

import (
	"context"
	"log"
	"math/rand"
	"time"
)

const (
	TimeoutLowerBound = 150 * time.Millisecond
	TimeoutRange      = 150 * time.Millisecond
)

func (rc *RaftCore) roleLoop() error {
	for {
		jitter := time.Duration(rand.Int63n(int64(TimeoutRange)))
		timeout := TimeoutLowerBound + jitter

		if rc.node.timer == nil {
			rc.node.timer = time.NewTimer(timeout)
		} else {
			rc.node.timer.Stop()
			rc.node.timer.Reset(timeout)
		}

		select {
		case <-rc.node.timer.C:
			if err := rc.startElection(); err != nil {
				log.Printf("[roleLoop] Error during startElection: %v\n", err)
				return err
			}
		case <-rc.node.heartbeatCh:
			continue
		case <-rc.node.stopCh:
			if rc.node.timer != nil {
				rc.node.timer.Stop()
			}
			return nil
		}
	}
}

func (n *node) tickHeartbeat() {
	select {
	case n.heartbeatCh <- struct{}{}:
	default:
		// so that it will not block RPC handler
	}
}

func (rc *RaftCore) startElection() error {
	rc.mu.Lock()
	curTerm, err := rc.node.getCurrentTerm()
	if err != nil {
		return err
	}
	newTerm := curTerm + 1
	if err := rc.node.setCurrentTerm(newTerm); err != nil {
		return err
	}

	if rc.node.state == Follower {
		rc.node.state = Candidate
		rc.receivedVotes++
		for _, peer := range rc.Peers {
			go rc.voteRequest(peer.netAddr, newTerm)
		}
	} else if rc.node.state == Candidate {
		rc.receivedVotes = 0 // empty previous term votes
		rc.receivedVotes++
		for _, peer := range rc.Peers {
			go rc.voteRequest(peer.netAddr, newTerm)
		}
	}
	rc.mu.Unlock()
	return nil
}

func (rc *RaftCore) becomeLeader(term uint32) {
	// caller must hold rc.mu
	log.Printf("[RaftCore] Node %s becomes Leader in term %d\n", rc.Info.id, term)

	rc.node.state = Leader
	size := len(rc.Peers) + 1

	rc.node.nextIndex = make(map[string]uint32, size)
	rc.node.matchIndex = make(map[string]uint32, size)
	for _, p := range rc.Peers {
		rc.node.nextIndex[p.id] = rc.node.lastLogIndex + 1
		rc.node.matchIndex[p.id] = 0
	}

	ctx, cancel := context.WithCancel(rc.node.leaderCtx)
	rc.node.leaderCancel = cancel

	go rc.heartbeatLoop(ctx, term)
	go rc.leaderStartServing(ctx)
}

func (rc *RaftCore) leaderStartServing(ctx context.Context) {
	for {
		log.Println("[leaderStartServing] Serving...")
		time.Sleep(time.Second * 2)
	}
}
