package timex_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/bingoohuang/gg/pkg/timex"
	"github.com/stretchr/testify/assert"
)

func TestUnmashalMsg(t *testing.T) {
	fmt.Println(time.Now().Format(time.RFC3339))
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

	assert.True(t, errors.Is(err, timex.ErrUnknownTimeFormat))

	assert.Equal(t, timex.JSONTime(time.Unix(0, 123*1000000)), msg.A)
	assert.Equal(t, timex.JSONTime(time.Unix(0, 123*1000000)), msg.F)

	assert.Equal(t, timex.JSONTime(zero), msg.O)
	assert.Equal(t, timex.JSONTime(p), msg.B)
	assert.Equal(t, timex.JSONTime(p), msg.C)
	assert.Equal(t, p, time.Time(msg.D).Local().Add(-8*time.Hour))
	assert.Equal(t, timex.JSONTime(p), msg.E)
	assert.Equal(t, time.Time(msg.D).Format("20060102150405"), "20200318105154")
}

type Msg struct {
	O timex.JSONTime
	A timex.JSONTime
	B timex.JSONTime
	C timex.JSONTime
	E timex.JSONTime
	F timex.JSONTime
	D timex.JSONTime `json:"d"`
	G timex.JSONTime
}
