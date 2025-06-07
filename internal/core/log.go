package core

import (
	"encoding/binary"
	"fmt"
)

type logEntry struct {
	logTerm  uint32
	logIndex uint32
	command  []byte
}

type Log []logEntry

func bytesToLogEntry(raw []byte) (*logEntry, error) {
	if len(raw) < 8 {
		return nil, fmt.Errorf("Need at least 8 bytes: term (4 Bytes) and index (4 Bytes)\n")
	}
	logTerm := binary.BigEndian.Uint32(raw[:4])
	logIndex := binary.BigEndian.Uint32(raw[4:8])
	command := raw[8:]
	return &logEntry{
		logTerm:  logTerm,
		logIndex: logIndex,
		command:  command,
	}, nil
}

func logEntryToKeyBytes(entry *logEntry, lastLogIndex uint64) ([]byte, []byte) {
	command := entry.command
	buf := make([]byte, 8+len(command))
	binary.BigEndian.PutUint32(buf[:4], entry.logTerm)
	binary.BigEndian.PutUint32(buf[4:8], entry.logIndex)
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
