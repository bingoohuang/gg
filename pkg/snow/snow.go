// Package snow provides a very simple Twitter snowflake generator and parser.
package snow

import (
	"fmt"
	"sync"
	"time"
)

// A Node struct holds the basic information needed for a snowflake generator node
type Node struct {
	mu sync.Mutex

	option    Option
	nodeID    int64
	stepMask  int64
	timeShift uint8
	nodeShift uint8

	time int64
	step int64

	// epoch is snowflake epoch in milliseconds.
	epoch    int64
	nodeMask int64

	epochTime time.Time
	unit      time.Duration
}

// GetOption return the option.
func (n *Node) GetOption() Option { return n.option }

// GetEpoch returns an int64 epoch is snowflake epoch in milliseconds.
func (n *Node) GetEpoch() int64 { return n.epoch }

// GetTime returns an int64 unix timestamp in milliseconds of the snowflake ID time.
func (n *Node) GetTime() int64 { return n.time }

// GetNodeID returns an int64 of the snowflake ID node number
func (n *Node) GetNodeID() int64 { return n.nodeID }

// GetStep returns an int64 of the snowflake step (or sequence) number
func (n *Node) GetStep() int64 { return n.step }

func (n *Node) applyOption(o Option) error {
	n.option = o
	n.nodeID = o.NodeID

	var nodeMax int64 = -1 ^ (-1 << o.NodeBits)

	n.nodeMask = nodeMax << o.StepBits
	n.stepMask = -1 ^ (-1 << o.StepBits)
	n.timeShift = uint8(o.NodeBits + o.StepBits)
	n.nodeShift = uint8(o.StepBits)

	curTime := time.Now()
	// add time.Duration to curTime to make sure we use the monotonic clock if available
	n.epochTime = curTime.Add(time.Unix(o.Epoch/1e3, (o.Epoch%1000)*1e6).Sub(curTime))

	n.unit = o.TimestampUnit / time.Millisecond
	if n.unit == 0 {
		n.unit = 1
	}
	n.epoch = int64(n.epochTime.Nanosecond() / 1e6)

	if o.NodeID >= 0 && o.NodeID <= nodeMax {
		return nil
	}

	return fmt.Errorf("NodeID %d must be between 0 and %d", o.NodeID, nodeMax)
}

// NewNode returns a new snowflake node that can be used to generate snowflake IDs.
func NewNode(optionFns ...OptionFn) (*Node, error) {
	option := Option{NodeBits: -1, StepBits: -1, NodeID: -1}
	option.Apply(optionFns...)

	n := &Node{}
	if err := n.applyOption(option); err != nil {
		return nil, err
	}

	return n, nil
}

// Next creates and returns a unique snowflake ID
// To help guarantee uniqueness
// - Make sure your system is keeping accurate system time
// - Make sure you never have multiple nodes running with the same node ID
func (n *Node) Next() ID {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.next()
}

func (n *Node) next() ID {
	now := n.now()
	if now == n.time {
		if n.step = (n.step + 1) & n.stepMask; n.step == 0 {
			for now <= n.time {
				n.sleep()
				now = n.now()
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	return ID(now<<n.timeShift | n.nodeID<<n.nodeShift | n.step)
}

func (n *Node) sleep() {
	if n.unit > 1 {
		time.Sleep(n.unit * time.Millisecond)
	}
}

func (n *Node) now() int64 {
	return time.Since(n.epochTime).Nanoseconds() / 1e6 / int64(n.unit)
}
