package util

import (
	"encoding/binary"
	"fmt"
)

func BytesToUint32(raw []byte) (uint32, error) {
	if len(raw) < 4 {
		return 0, fmt.Errorf("need at least 4 bytes, got %d", len(raw))
	}
	u := binary.BigEndian.Uint32(raw[:4])
	return u, nil
}
