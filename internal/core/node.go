package core

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/jialuohu/curlyraft"
	"github.com/jialuohu/curlyraft/internal/clog"
	"github.com/jialuohu/curlyraft/internal/persistence"
	"log"
	"time"
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
	nextIndex          map[string]uint32
	matchIndex         map[string]uint32
	leaderCtx          context.Context
	leaderBecomeCtx    context.Context
	leaderCancel       context.CancelFunc
	leaderBecomeCancel context.CancelFunc
}

func newNode(storageDir string, sm curlyraft.StateMachine) *node {
	n := &node{
		state:       Follower,
		heartbeatCh: make(chan struct{}, 1),
		stopCh:      make(chan struct{}),
		sm:          sm,
		storage:     persistence.NewStorage(storageDir),
	}

	if err := n.setCurrentTerm(InitialTerm); err != nil {
		log.Fatalf("%s Failed to initialize term value into storage: %v\n", clog.CRedNode("newNode"), err)
	}
	if err := n.setVotedFor(VotedForNoOne); err != nil {
		log.Fatalf("%s Failed to initialize votedFor value into storage: %v\n", clog.CRedNode("newNode"), err)
	}

	key := keyForIndex(dummyLogEntry.logIndex)
	val := marshalEntry(&dummyLogEntry)
	if err := n.storage.Set(key, val); err != nil {
		log.Fatalf("%s Failed to initialize log: %v\n", clog.CRedNode("newNode"), err)
	}

	//n.leaderCtx, n.leaderCancel = context.WithCancel(context.Background())
	//n.leaderBecomeCtx, n.leaderBecomeCancel = context.WithCancel(context.Background())

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

	raw, closer, found, err := n.storage.Get([]byte("CurrentTerm"))
	if !found && err != nil {
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
	raw, closer, found, err := n.storage.Get([]byte("VotedFor"))
	if !found && err != nil {
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
		key := iter.Key()
		raw, err := iter.ValueAndErr()
		if err != nil {
			log.Printf("[GetLog] Failed to get one of value from iterator: %v\n", err)
			return nil, err
		}

		idxBytes := key[len(LogPrefix):]
		logIndex := binary.BigEndian.Uint32(idxBytes)
		entry, err := bytesToLogEntry(raw, logIndex)
		if err != nil {
			log.Printf("[GetLog] Failed to parse raw log entry: %v\n", err)
			return nil, err
		}

		out = append(out, entry)
	}
	return out, nil
}

func (n *node) appendEntry(entry *logEntry) error {
	log.Printf("%s Start append an entry:logTerm=%d,logIndex=%d\n",
		clog.CGreenNode("appendEntry"), entry.logTerm, entry.logIndex)
	key := keyForIndex(entry.logIndex)
	val := marshalEntry(entry)
	if err := n.storage.Set(key, val); err != nil {
		log.Printf("%s Failed to store entry into database: %v\n", clog.CRedNode("appendEntry"), err)
		return err
	}
	return nil
}

func (n *node) getLogTerm(idx uint32) (uint32, bool) {
	raw, closer, found, err := n.storage.Get(keyForIndex(idx))
	if !found && err != nil {
		log.Printf("%s Got error while getting key: %s : %v\n", clog.CRedNode("getLogTerm"), keyForIndex(idx), err)
		return 0, false
	}
	defer func() {
		if closer == nil {
			return
		}
		if err := closer.Close(); err != nil {
			log.Printf("[getLogTerm] Failed to close storage: %v\n", err)
		}
	}()

	if found {
		entry, err := bytesToLogEntry(raw, idx)
		if err != nil {
			log.Printf("%s Got error while doing bytesToLogEntry: %v\n", clog.CRedNode("getLogTerm"), err)
			return 0, false
		}

		return entry.logTerm, true
	} else {
		return 0, false
	}
}

func (n *node) deleteConflicts(fromIndex uint32) error {
	log.Printf("%s Start delete conflicts\n", clog.CGreenNode("deleteConflicts"))
	startKey := keyForIndex(fromIndex)
	endKey := append([]byte(LogPrefix), 0xFF)
	it, err := n.storage.NewIter(&persistence.IterOptions{
		LowerBound: startKey,
		UpperBound: endKey,
	})
	if err != nil {
		log.Printf("%s Iterator error: %v", clog.CRedNode("deleteConflicts"), err)
		return err
	}
	defer func() {
		if err := it.Close(); err != nil {
			log.Printf("%s Failed to close iterator: %v\n", clog.CRedNode("deleteConflicts"), err)
		}
	}()

	log.Printf("%s Looking for the fromIndex:%d\n", clog.CGreenNode("deleteConflicts"), fromIndex)
	for ok := it.First(); ok; ok = it.Next() {
		key := it.Key()
		if err := n.storage.Delete(key); err != nil {
			log.Printf("%s Failed to delete %s: %v", clog.CRedNode("deleteConflicts"), key, err)
			return err
		}
	}

	log.Printf("%s Conflicts resolved\n", clog.CGreenNode("deleteConflicts"))
	return nil
}

func (n *node) applyEntries() {
	log.Printf("%s Start apply entires\n", clog.CGreenNode("applyEntries"))
	for {
		next := n.lastApplied + 1
		log.Printf("%s Log entry:%d is going to be applied\n", clog.CGreenNode("applyEntries"), next)
		if next > n.commitIndex {
			log.Printf("%s The next:%d is over current node commitIndex:%d!\n",
				clog.CBlueNode("applyEntries"), next, n.commitIndex)
			return
		}

		// read that entry
		log.Printf("%s Read the bytes of entry\n", clog.CGreenNode("applyEntries"))
		raw, closer, found, err := n.storage.Get(keyForIndex(next))
		if !found && err != nil {
			log.Printf("%s Get entry %d: %v\n", clog.CRedNode("applyEntries"), next, err)
			return
		}
		data := make([]byte, len(raw))
		copy(data, raw)
		if err := closer.Close(); err != nil {
			log.Printf("%s Failed to close %v\n", clog.CRedNode("applyEntries"), err)
			return
		}

		log.Printf("%s Transform the bytes into entry struct\n", clog.CGreenNode("applyEntries"))
		entry, err := bytesToLogEntry(data, next)
		if err != nil {
			log.Printf("%s Parse entry %d: %v\n", clog.CRedNode("applyEntries"), next, err)
			return
		}

		// now apply to your state machine
		log.Printf("%s State machine applies the command\n", clog.CGreenNode("applyEntries"))
		if _, err := n.sm.Apply(entry.command); err != nil {
			log.Printf("%s State machine error at %d: %v", clog.CRedNode("applyEntries"), next, err)
		}

		log.Printf("%s Mark the log entry:%d applied\n", clog.CBlueNode("applyEntries"), n.lastApplied)
		n.lastApplied = next
	}
}

func (n *node) stopNode() {
	close(n.heartbeatCh)
	close(n.stopCh)
	if n.leaderCancel != nil {
		n.leaderCancel()
	}
}
