package core

import (
	"encoding/binary"
	"fmt"
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
	//sm          *curlyraft.StateMachine

	// persistence
	lastLogIndex uint32
	lastLogTerm  uint32
	storage      *persistence.Storage

	// leader only
	nextIndex  []uint32
	matchIndex []uint32
}

func newNode() *node {
	n := &node{
		mu:           sync.Mutex{},
		state:        Follower,
		commitIndex:  0,
		lastApplied:  0,
		timer:        nil,
		heartbeatCh:  make(chan struct{}, 1),
		stopCh:       make(chan struct{}),
		lastLogIndex: 0,
		lastLogTerm:  0,
		storage:      persistence.NewStorage(),
		nextIndex:    nil,
		matchIndex:   nil,
	}

	if err := n.setCurrentTerm(InitialTerm); err != nil {
		log.Fatalf("[newNode] Failed to initialize term value into storage: %v\n", err)
	}
	if err := n.setVotedFor(VotedForNoOne); err != nil {
		log.Fatalf("[newNode] Failed to initialize votedFor value into storage: %v\n", err)
	}

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

func (n *node) stopNode() {
	close(n.heartbeatCh)
	close(n.stopCh)
}
