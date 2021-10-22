package snow

import (
	"github.com/bingoohuang/gg/pkg/goip"
	"time"
)

// Option for the snowflake
type Option struct {
	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds
	// You may customize this to set a different epoch for your application.
	Epoch int64 // 1288834974657

	// NodeBits holds the number of bits to use for Node
	// Remember, you have a total 22 bits to share between Node/Step
	NodeBits int8 // 10

	// StepBits holds the number of bits to use for Step
	// Remember, you have a total 22 bits to share between Node/Step
	StepBits int8 // 12

	// NodeID for the snowflake.
	NodeID int64

	// TimestampUnit for the time goes unit, default is 1ms.
	TimestampUnit time.Duration
}

// Apply applies the option functions to the option.
// nolint gomnd
func (o *Option) Apply(fns ...OptionFn) {
	for _, fn := range fns {
		fn(o)
	}

	if o.Epoch == 0 {
		o.Epoch = 1288834974657
	}

	if o.NodeBits < 0 {
		o.NodeBits = 10
	}

	if o.StepBits < 0 {
		o.StepBits = 12
	}

	if o.NodeID < 0 {
		o.NodeID = defaultIPNodeID()
		var nodeMax int64 = -1 ^ (-1 << o.NodeBits)
		o.NodeID &= nodeMax
	}
}

// OptionFn defines the function prototype to apply options.
type OptionFn func(*Option)

// WithNodeID set the customized nodeID.
func WithNodeID(nodeID int64) OptionFn { return func(o *Option) { o.NodeID = nodeID } }

// WithNodeIDLocalIP set the customized nodeID  with the last 8 bits of local IP v4 and first 2 bits of p.
func WithNodeIDLocalIP(p int64, ip string) OptionFn {
	if ip == "" {
		ip, _ = goip.MainIP()
	}

	return func(o *Option) { o.NodeID = (p << 8) | ipNodeID(ip) }
}

// WithEpoch set the customized epoch.
func WithEpoch(epoch int64) OptionFn { return func(o *Option) { o.Epoch = epoch } }

// WithNodeBits set the customized NodeBits n.
func WithNodeBits(n int8) OptionFn { return func(o *Option) { o.NodeBits = n } }

// WithStepBits set the customized StepBits n.
func WithStepBits(n int8) OptionFn { return func(o *Option) { o.StepBits = n } }

// WithTimestampUnit set the customized TimestampUnit n.
func WithTimestampUnit(n time.Duration) OptionFn { return func(o *Option) { o.TimestampUnit = n } }
