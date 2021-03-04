package jsontime_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/bingoohuang/gg/pkg/jsontime"
	"github.com/stretchr/testify/assert"
)

func TestUnmashalMsg(t *testing.T) {
	p, _ := time.ParseInLocation("2006-01-02 15:04:05.000", "2020-03-18 10:51:54.198", time.Local)

	j := `{
		"O": "",
		"A": "123",
		"F": 123,
		"B": "2020-03-18 10:51:54.198",
		"C": "2020-03-18 10:51:54,198",
		"E": "2020-03-18T10:51:54,198",
		"d": "2020-03-18T10:51:54.198000Z",
		"G": "XYZ"
	}`

	var (
		zero time.Time
		msg  Msg
	)

	err := json.Unmarshal([]byte(j), &msg)

	assert.True(t, errors.Is(err, jsontime.ErrUnknownTimeFormat))

	assert.Equal(t, jsontime.Time(time.Unix(0, 123*1000000)), msg.A)
	assert.Equal(t, jsontime.Time(time.Unix(0, 123*1000000)), msg.F)

	assert.Equal(t, jsontime.Time(zero), msg.O)
	assert.Equal(t, jsontime.Time(p), msg.B)
	assert.Equal(t, jsontime.Time(p), msg.C)
	assert.Equal(t, jsontime.Time(p), msg.D)
	assert.Equal(t, jsontime.Time(p), msg.E)
	assert.Equal(t, time.Time(msg.D).Format("20060102150405"), "20200318105154")
}

type Msg struct {
	O jsontime.Time
	A jsontime.Time
	B jsontime.Time
	C jsontime.Time
	E jsontime.Time
	F jsontime.Time
	D jsontime.Time `json:"d"`
	G jsontime.Time
}
