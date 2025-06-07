package core

import (
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/internal/persistence"
	"github.com/jialuohu/curlyraft/internal/util"
	"log"
)

type nodeInfo struct {
	id      string
	netAddr string
}

type node struct {
	// node info
	state       Role
	commitIndex uint32
	lastApplied uint32
	sm          *curlyraft.StateMachine

	// persistence
	lastLogIndex uint64
	storage      *persistence.Storage

	// leader only
	nextIndex  []uint32
	matchIndex []uint32
}

func newNode() *node {
	return &node{
		state:        Follower,
		commitIndex:  0,
		lastApplied:  0,
		lastLogIndex: 0,
		storage:      persistence.NewStorage(),
		nextIndex:    nil,
		matchIndex:   nil,
	}
}

func (n *node) getCurrentTerm() (uint32, error) {
	raw, closer, err := n.storage.Get([]byte("CurrentTerm"))
	if err != nil {
		log.Printf("[GetCurrentTerm] Failed to get value from storage: %v\n", err)
		return InvalidTerm, err
	}
	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("[GetCurrentTerm] Failed to close storage cache access: %v\n", err)
		}
	}()

	curTermBytes := make([]byte, len(raw))
	copy(curTermBytes, raw)
	curTerm, err := util.BytesToUint32(curTermBytes)
	if err != nil {
		log.Printf("[GetCurrentTerm] Failed to convert bytes into int32: %v\n", err)
		return InvalidTerm, err
	}
	return curTerm, nil
}

func (n *node) getVotedFor() (uint32, error) {
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

	votedForBytes := make([]byte, len(raw))
	copy(votedForBytes, raw)
	votedFor, err := util.BytesToUint32(votedForBytes)
	if err != nil {
		log.Printf("[GetVotedFor] Failed to convert bytes into int32: %v\n", err)
		return VotedForNoOne, err
	}
	return votedFor, nil
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
	return nil
}
