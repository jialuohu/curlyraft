package core

import (
	"encoding/binary"
	"fmt"
)

type LogEntry struct {
	LogTerm  uint32
	LogIndex uint32
	Command  []byte
}

type Log []LogEntry

func BytesToLogEntry(raw []byte) (*LogEntry, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("Need at least 8 bytes: term (4 Bytes) and index (4 Bytes)\n")
	}
	logTerm := binary.BigEndian.Uint32(raw[:4])
	logIndex := binary.BigEndian.Uint32(raw[4:8])
	command := raw[8:]
	return &LogEntry{
		LogTerm:  logTerm,
		LogIndex: logIndex,
		Command:  command,
	}, nil
}

func LogEntryToKeyBytes(entry *LogEntry, lastLogIndex uint64) ([]byte, []byte) {
	command := entry.Command
	buf := make([]byte, 8+len(command))
	binary.BigEndian.PutUint32(buf[:4], entry.LogTerm)
	binary.BigEndian.PutUint32(buf[4:8], entry.LogIndex)
	copy(buf[8:], command)

	key := make([]byte, len(LogPrefix)+8)
	copy(key, LogPrefix)
	binary.BigEndian.AppendUint64(key, lastLogIndex)
	return key, buf
}

const (
	InvalidTerm   = 0
	VotedForNoOne = 0
	LogPrefix     = "log/"
)
