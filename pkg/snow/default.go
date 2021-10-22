package snow

import "time"

// DefaultNode is the global default snowflake node object.
// nolint gochecknoglobals
var DefaultNode, _ = NewNode()

// GetOption return the option.
func GetOption() Option { return DefaultNode.option }

// GetEpoch returns an int64 epoch is snowflake epoch in milliseconds.
func GetEpoch() int64 { return DefaultNode.epoch }

// GetTime returns an int64 unix timestamp in milliseconds of the snowflake ID time.
func GetTime() int64 { return DefaultNode.time }

// GetNodeID returns an int64 of the snowflake ID node number
func GetNodeID() int64 { return DefaultNode.nodeID }

// GetStep returns an int64 of the snowflake step (or sequence) number
func GetStep() int64 { return DefaultNode.step }

// Next creates and returns a unique snowflake ID
// To help guarantee uniqueness
// - Make sure your system is keeping accurate system time
// - Make sure you never have multiple nodes running with the same node ID
func Next() ID { return DefaultNode.Next() }

var DefaultNode32, _ = NewNode(WithNodeBits(2), WithStepBits(1), WithTimestampUnit(1*time.Second))

// Next32 creates and returns a unique snowflake ID for positive int32.
// only for low frequency usages.
// unsigned(1) + timestamp(28) + node ID(2) + step(1)
// can use 2^28/60/60/24/365 ≈ 8.5 年
func Next32() ID { return DefaultNode32.Next() }
