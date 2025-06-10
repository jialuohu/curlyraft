package core

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/internal/clog"
	"github.com/jialuohu/curlyraft/internal/persistence"
	"log"
	"sync"
	"time"
)

type nodeInfo struct {
	id      string
	netAddr string
}

type node struct {
	mu sync.Mutex

	// node info
	state       Role
	commitIndex uint32
	lastApplied uint32
	timer       *time.Timer
	heartbeatCh chan struct{}
	stopCh      chan struct{}
	leaderId    *nodeInfo
	sm          curlyraft.StateMachine

	// persistence
	lastLogIndex uint32
	lastLogTerm  uint32
	storage      *persistence.Storage

	// leader only
	nextIndex    map[string]uint32
	matchIndex   map[string]uint32
	leaderCtx    context.Context
	leaderCancel context.CancelFunc
}

func newNode(storageDir string, sm curlyraft.StateMachine) *node {
	n := &node{
		mu:           sync.Mutex{},
		state:        Follower,
		commitIndex:  0,
		lastApplied:  0,
		timer:        nil,
		heartbeatCh:  make(chan struct{}, 1),
		stopCh:       make(chan struct{}),
		leaderId:     nil,
		sm:           sm,
		lastLogIndex: 0,
		lastLogTerm:  0,
		storage:      persistence.NewStorage(storageDir),
		nextIndex:    nil,
		matchIndex:   nil,
		leaderCtx:    context.Background(),
		leaderCancel: nil,
	}

	if err := n.setCurrentTerm(InitialTerm); err != nil {
		log.Fatalf("%s Failed to initialize term value into storage: %v\n", clog.CRedNode("newNode"), err)
	}
	if err := n.setVotedFor(VotedForNoOne); err != nil {
		log.Fatalf("%s Failed to initialize votedFor value into storage: %v\n", clog.CRedNode("newNode"), err)
	}

	//if err := n.storage.Set([]byte(LogPrefix), []byte("")); err != nil {
	//	log.Fatalf("%s Failed to initialize log: %v\n", clog.CRedNode("newNode"), err)
	//}

	return n
}

func (n *node) getCurrentTerm() (uint32, error) {
	bytesToUint32 := func(raw []byte) (uint32, error) {
		if len(raw) < 4 {
			return 0, fmt.Errorf("need at least 4 bytes, got %d", len(raw))
		}
		u := binary.BigEndian.Uint32(raw[:4])
		return u, nil
	}

	raw, closer, err := n.storage.Get([]byte("CurrentTerm"))
	if err != nil {
		log.Printf("[GetCurrentTerm] Failed to get value from storage: %v\n", err)
		return InitialTerm, err
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("[GetCurrentTerm] Failed to close storage cache access: %v\n", err)
		}
	}()

	curTermBytes := make([]byte, len(raw))
	copy(curTermBytes, raw)
	curTerm, err := bytesToUint32(curTermBytes)
	if err != nil {
		log.Printf("[GetCurrentTerm] Failed to convert bytes into int32: %v\n", err)
		return InitialTerm, err
	}
	return curTerm, nil
}

func (n *node) setCurrentTerm(term uint32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, term)
	if err := n.storage.Set([]byte("CurrentTerm"), buf); err != nil {
		log.Printf("[SetCurrentTerm] Failed to set term value into storage: %v\n", err)
		return err
	}
	return nil
}

func (n *node) getVotedFor() (string, error) {
	raw, closer, err := n.storage.Get([]byte("VotedFor"))
	if err != nil {
		log.Printf("[GetVotedFor] Failed to get value from storage: %v\n", err)
		return VotedForNoOne, err
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("[GetVotedFor] Failed to close storage cache access: %v\n", err)
		}
	}()

	return string(raw), nil
}

func (n *node) setVotedFor(votedFor string) error {
	if err := n.storage.Set([]byte("VotedFor"), []byte(votedFor)); err != nil {
		log.Printf("[SetVotedFor] Failed to set votedFor value into storage: %v\n", err)
		return err
	}
	return nil
}

func (n *node) getLog() ([]*logEntry, error) {
	lb := []byte(LogPrefix)
	ub := append(lb, 0xFF)
	iter, err := n.storage.NewIter(
		&persistence.IterOptions{
			LowerBound: lb,
			UpperBound: ub,
		})
	if err != nil {
		log.Printf("[GetLog] Failed to open database iterator: %v\n", err)
		return nil, err
	}
	defer func() {
		if err := iter.Close(); err != nil {
			log.Printf("[GetLog] Failed to close database iterater: %v\n", err)
		}
	}()

	out := make([]*logEntry, 0)
	for ok := iter.First(); ok; ok = iter.Next() {
		raw, err := iter.ValueAndErr()
		if err != nil {
			log.Printf("[GetLog] Failed to get one of value from iterator: %v\n", err)
			return nil, err
		}

		entry, err := bytesToLogEntry(raw)
		if err != nil {
			log.Printf("[GetLog] Failed to parse raw log entry: %v\n", err)
			return nil, err
		}

		out = append(out, entry)
	}
	return out, nil
}

func (n *node) appendEntry(entry *logEntry) error {
	logIndex := n.lastLogIndex + 1
	key, val := logEntryToKeyBytes(entry, n.lastLogIndex)
	if err := n.storage.Set(key, val); err != nil {
		log.Printf("[AppendEntry] Failed to store entry into database: %v\n", err)
		return err
	}
	n.lastLogIndex = logIndex
	if n.lastLogTerm < entry.logTerm {
		n.lastLogTerm = entry.logTerm
	}
	return nil
}

func (n *node) getLogTerm(idx uint32) (uint32, bool) {
	if idx == 0 {
		log.Printf("%s idx is 0, return 0, true\n", clog.CGreenNode("getLogTerm"))
		return 0, true
	}
	raw, closer, err := n.storage.Get(keyForIndex(idx))
	if err != nil {
		log.Printf("%s Got error while getting key: %s : %v\n", clog.CRedNode("getLogTerm"), keyForIndex(idx), err)
		return 0, false
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("[getLogTerm] Failed to close storage: %v\n", err)
		}
	}()

	entry, err := bytesToLogEntry(raw)
	if err != nil {
		log.Printf("%s Got error while doing bytesToLogEntry: %v\n", clog.CRedNode("getLogTerm"), err)
		return 0, false
	}

	return entry.logTerm, true
}

func (n *node) stopNode() {
	close(n.heartbeatCh)
	close(n.stopCh)
	if n.leaderCancel != nil {
		n.leaderCancel()
	}
}

func (n *node) deleteConflicts(fromIndex uint32) error {
	startKey := keyForIndex(fromIndex)
	endKey := append([]byte(LogPrefix), 0xFF)
	it, err := n.storage.NewIter(&persistence.IterOptions{
		LowerBound: startKey,
		UpperBound: endKey,
	})
	if err != nil {
		log.Printf("[deleteConflicts] Iterator error: %v", err)
		return err
	}
	defer func() {
		if err := it.Close(); err != nil {
			log.Printf("[deleteConflicts] Failed to close iterator: %v", err)
		}
	}()

	for ok := it.First(); ok; ok = it.Next() {
		key := it.Key()
		if err := n.storage.Delete(key); err != nil {
			log.Printf("[deleteConflicts] Failed to delete %s: %v", key, err)
			return err
		}
	}

	return nil
}

func (n *node) applyEntries() {
	for {
		next := n.lastApplied + 1
		if next > n.commitIndex {
			return
		}

		// read that entry
		raw, closer, err := n.storage.Get(keyForIndex(next))
		if err != nil {
			log.Printf("[applyEntries] Get entry %d: %v\n", next, err)
			return
		}
		data := make([]byte, len(raw))
		copy(data, raw)
		if err := closer.Close(); err != nil {
			log.Printf("[applyEntries] Failed to close %v\n", err)
			return
		}

		entry, err := bytesToLogEntry(data)
		if err != nil {
			log.Printf("[applyEntries] Parse entry %d: %v\n", next, err)
			return
		}

		// now apply to your state machine
		if _, err := n.sm.Apply(entry.command); err != nil {
			log.Printf("[applyEntries] state machine error at %d: %v", next, err)
		}

		// mark it applied
		n.lastApplied = next
	}
}
