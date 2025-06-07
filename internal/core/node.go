package core

import (
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/internal/persistence"
	"github.com/jialuohu/curlyraft/internal/util"
	"log"
)

type Node struct {
	// node info
	State       Role
	CommitIndex uint32
	LastApplied uint32
	Cfg         *NodeCfg
	Sm          *curlyraft.StateMachine

	// persistence
	LastLogIndex uint64
	storage      *persistence.Storage

	// leader only
	NextIndex  []uint32
	MatchIndex []uint32
}

func NewNode(cfg *NodeCfg, sm *curlyraft.StateMachine) *Node {
	go NewRpcServer(cfg.Addr)
	return &Node{
		State:        Follower,
		CommitIndex:  0,
		LastApplied:  0,
		Cfg:          cfg,
		Sm:           sm,
		LastLogIndex: 0,
		storage:      persistence.NewStorage(),
		NextIndex:    nil,
		MatchIndex:   nil,
	}
}

func (n *Node) GetCurrentTerm() (uint32, error) {
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

func (n *Node) GetVotedFor() (uint32, error) {
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

func (n *Node) GetLog() ([]*LogEntry, error) {
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

	out := make([]*LogEntry, 0)
	for ok := iter.First(); ok; ok = iter.Next() {
		raw, err := iter.ValueAndErr()
		if err != nil {
			log.Printf("[GetLog] Failed to get one of value from iterator: %v\n", err)
			return nil, err
		}

		entry, err := BytesToLogEntry(raw)
		if err != nil {
			log.Printf("[GetLog] Failed to parse raw log entry: %v\n", err)
			return nil, err
		}

		out = append(out, entry)
	}
	return out, nil
}

func (n *Node) AppendEntry(entry *LogEntry) error {
	logIndex := n.LastLogIndex + 1
	key, val := LogEntryToKeyBytes(entry, n.LastLogIndex)
	if err := n.storage.Set(key, val); err != nil {
		log.Printf("[AppendEntry] Failed to store entry into database: %v\n", err)
		return err
	}
	n.LastLogIndex = logIndex
	return nil
}
