package snow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// nolint gomnd
func TestIpNodeID(t *testing.T) {
	that := assert.New(t)
	that.Equal(int64(10), ipNodeID("192.168.1.10"))
	that.Equal(int64(11), ipNodeID("192.168.1.11"))
}
