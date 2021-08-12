// Package snow provides a very simple Twitter snowflake generator and parser.
package snow

import (
	"errors"
	"strconv"
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
	timeFn    func() int64

	time int64
	step int64

	// epoch is snowflake epoch in milliseconds.
	epoch    int64
	nodeMask int64
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

// nolint gomnd
func (n *Node) applyOption(o Option) error {
	n.option = o
	n.nodeID = o.NodeID

	var (
		nodeMax int64 = -1 ^ (-1 << o.NodeBits)
	)

	n.nodeMask = nodeMax << o.StepBits
	n.stepMask = -1 ^ (-1 << o.StepBits)
	n.timeShift = o.NodeBits + o.StepBits
	n.nodeShift = o.StepBits

	curTime := time.Now()
	// add time.Duration to curTime to make sure we use the monotonic clock if available
	epoch := curTime.Add(time.Unix(o.Epoch/1000, (o.Epoch%1000)*1000000).Sub(curTime))

	n.timeFn = func() int64 { return time.Since(epoch).Nanoseconds() / 1000000 /* nolint gomnd*/ }
	n.epoch = int64(epoch.Nanosecond() / 1000000)

	if o.NodeID >= 0 && o.NodeID <= nodeMax {
		return nil
	}

	return errors.New("NodeID must be between 0 and " + strconv.FormatInt(nodeMax, 10))
}

// NewNode returns a new snowflake node that can be used to generate snowflake IDs.
func NewNode(optionFns ...OptionFn) (*Node, error) {
	option := Option{NodeID: -1}
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
	now := n.timeFn()
	if now == n.time {
		if n.step = (n.step + 1) & n.stepMask; n.step == 0 {
			for now <= n.time {
				now = n.timeFn()
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	return ID(now<<n.timeShift | n.nodeID<<n.nodeShift | n.step)
}
