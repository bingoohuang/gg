package ss_test

import (
	"testing"

	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/stretchr/testify/assert"
)

func TestJoinMap(t *testing.T) {
	assert.Equal(t, "svc=braft", ss.JoinMap(map[string]string{"svc": "braft"}, "=", ","))
}
