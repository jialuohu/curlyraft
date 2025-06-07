package core

import (
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
			//TODO implement me
		case <-rc.node.stopCh:
			if rc.node.timer != nil {
				rc.node.timer.Stop()
			}
			return nil
		}
	}
}

// TODO implement me
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
	log.Printf("[RaftCore] Node %s becomes Leader in term %d\n",
		rc.Info.id, term)

	rc.node.state = Leader
	size := len(rc.Peers) + 1
	rc.node.nextIndex = make([]uint32, size)
	rc.node.matchIndex = make([]uint32, size)

	go rc.heartbeatLoop()
}
