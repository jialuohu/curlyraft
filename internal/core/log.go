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

var dummyLogEntry = logEntry{
	logTerm:  0,
	logIndex: 0,
	command:  nil,
}

func bytesToLogEntry(raw []byte, logIndex uint32) (*logEntry, error) {
	if len(raw) < 4 {
		return nil, fmt.Errorf("Need at least 4 bytes: term (4 Bytes) but only has %d\n", len(raw))
	}
	logTerm := binary.BigEndian.Uint32(raw[:4])
	command := raw[4:]
	return &logEntry{
		logTerm:  logTerm,
		logIndex: logIndex,
		command:  command,
	}, nil
}

func marshalEntry(entry *logEntry) []byte {
	cmd := entry.command
	term := entry.logTerm
	buf := make([]byte, len(LogPrefix)+len(cmd))
	binary.BigEndian.PutUint32(buf[:4], term)
	copy(buf[4:], cmd)
	return buf
}

func keyForIndex(idx uint32) []byte {
	b := make([]byte, len(LogPrefix)+4)
	copy(b, LogPrefix)
	binary.BigEndian.PutUint32(b[len(LogPrefix):], idx)
	return b
}

const (
	InitialTerm   = 0
	VotedForNoOne = "none"
	LogPrefix     = "log/"
)
