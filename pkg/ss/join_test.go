package ss_test

import (
	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJoinMap(t *testing.T) {
	assert.Equal(t, "svc=braft", ss.JoinMap(map[string]string{"svc": "braft"}, "=", ","))
}
