package idgenerator

import (
	"fmt"
	"sync"
	"time"
)

const (
	customEpochMillis = int64(1704067200000) // 2024-01-01T00:00:00Z
	nodeBits          = uint(10)
	sequenceBits      = uint(12)
	maxNodeID         = int64(-1 ^ (-1 << nodeBits))
	sequenceMask      = int64(-1 ^ (-1 << sequenceBits))
)

type SnowflakeGenerator struct {
	mu            sync.Mutex
	nodeID        int64
	lastTimestamp int64
	sequence      int64
}

func NewSnowflakeGenerator(nodeID int64) (*SnowflakeGenerator, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, fmt.Errorf("snowflake node id must be between 0 and %d", maxNodeID)
	}

	return &SnowflakeGenerator{nodeID: nodeID}, nil
}

func (g *SnowflakeGenerator) NewID() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := currentMillis()
	if timestamp < g.lastTimestamp {
		timestamp = g.lastTimestamp
	}

	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & sequenceMask
		if g.sequence == 0 {
			timestamp = g.waitNextMillis(timestamp)
		}
	} else {
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	snowflakeTimestamp := timestamp - customEpochMillis
	id := (snowflakeTimestamp << (nodeBits + sequenceBits)) |
		(g.nodeID << sequenceBits) |
		g.sequence

	return id, nil
}

func (g *SnowflakeGenerator) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := currentMillis()
	for timestamp <= lastTimestamp {
		time.Sleep(time.Millisecond)
		timestamp = currentMillis()
	}

	return timestamp
}

func currentMillis() int64 {
	return time.Now().UnixMilli()
}
