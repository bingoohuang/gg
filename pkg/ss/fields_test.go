package ss_test

import (
	"testing"

	"github.com/bingoohuang/gg/pkg/ss"
	"github.com/stretchr/testify/assert"
)

func TestFieldsX(t *testing.T) {
	assert.Nil(t, ss.FieldsX("(a b) c", "(", ")", 0))
	assert.Equal(t, []string{"(a b)  c"}, ss.FieldsX("(a b)  c ", "(", ")", 1))
	assert.Equal(t, []string{"(a b)", "c"}, ss.FieldsX("(a b)  c", "(", ")", 2))
	assert.Equal(t, []string{"(a b)", "c  d e"}, ss.FieldsX("(a b)  c  d e ", "(", ")", 2))
	assert.Equal(t, []string{"(a b)", "c"}, ss.FieldsX("(a b) c", "(", ")", -1))
	assert.Equal(t, []string{"(a b)", "(c d)"}, ss.FieldsX(" (a b) (c d) ", "(", ")", -1))
	assert.Equal(t, []string{"(中 华) (人 民)"}, ss.FieldsX("(中 华) (人 民)  ", "(", ")", 1))
	assert.Equal(t, []string{"(中 华)", "(人 民)"}, ss.FieldsX(" (中 华) (人 民)  ", "(", ")", -1))
	assert.Equal(t, []string{"(中 华)", "(人 民)  共和国"}, ss.FieldsX(" (中 华) (人 民)  共和国", "(", ")", 2))
}

func TestFields(t *testing.T) {
	assert.Nil(t, ss.FieldsN("a b c", 0), nil)
	assert.Equal(t, []string{"a b c"}, ss.FieldsN(" a b c ", 1))
	assert.Equal(t, ss.FieldsN(" a b c", 2), []string{"a", "b c"})
	assert.Equal(t, ss.FieldsN("a   b c", 3), []string{"a", "b", "c"})
	assert.Equal(t, ss.FieldsN("a b c", 4), []string{"a", "b", "c"})
	assert.Equal(t, ss.FieldsN("a b c", -1), []string{"a", "b", "c"})
	assert.Equal(t, []string{"中国", "c"}, ss.FieldsN("中国 c", -1))
	assert.Equal(t, []string{"中国 c"}, ss.FieldsN("中国 c", 1))
	assert.Equal(t, []string{"中国", "人民  共和国"}, ss.FieldsN("   中国 人民  共和国   ", 2))
	assert.Equal(t, []string{"中国", "人民共和国"}, ss.FieldsN("   中国  人民共和国  ", 2))
	assert.Equal(t, []string{"中国", "人民共和国"}, ss.FieldsN("  中国  人民共和国  ", 3))
}
