package randx

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"log"
	"math"
	"math/big"
	"time"
)

// from https://github.com/thanhpk/randstr

// list of default letters that can be used to make a random string when calling String
// function with no letters provided
var defLetters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// String generates a random string using only letters provided in the letters parameter
// if user ommit letters parameters, this function will use defLetters instead
func String(n int, letters ...string) string {
	var letterRunes []rune
	if len(letters) == 0 {
		letterRunes = defLetters
	} else {
		letterRunes = []rune(letters[0])
	}

	var bb bytes.Buffer
	bb.Grow(n)
	l := uint32(len(letterRunes))
	// on each loop, generate one random rune and append to output
	for i := 0; i < n; i++ {
		bb.WriteRune(letterRunes[binary.BigEndian.Uint32(Bytes(4))%l])
	}
	return bb.String()
}

// Bytes generates n random bytes.
func Bytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

var rander = rand.Reader // random function

func Time() time.Time {
	min := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC)
	max := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC)
	return TimeBetween(min, max)
}

func TimeBetween(min, max time.Time) time.Time {
	minUnit, maxUnix := min.Unix(), max.Unix()
	n, _ := rand.Int(rander, big.NewInt(maxUnix-minUnit))
	return time.Unix(n.Int64()+minUnit, 0)
}

func Bool() bool { return Int64Between(0, 1) == 0 }

func Int64() int64 { return Int64N(math.MaxInt64) }

func Int64N(n int64) int64 {
	v, _ := rand.Int(rander, big.NewInt(n))
	return v.Int64()
}

func Int64Between(min, max int64) (v int64) { return Int64N(max-min+1) + min }

func IntN(n int) int {
	v, _ := rand.Int(rander, big.NewInt(int64(n)))
	return int(v.Int64())
}
func Int() int { return int(Int32()) }

func IntBetween(min, max int) int { return IntN(max-min+1) + min }

func Int32N(n int) int32 {
	v, _ := rand.Int(rander, big.NewInt(int64(n)))
	return int32(v.Int64())
}

func Int32() int32 { return Int32N(math.MaxInt32) }

func Uint64N(n int64) uint64 {
	v, _ := rand.Int(rander, big.NewInt(n))
	return v.Uint64()
}

func Uint64() (v uint64) {
	err := binary.Read(rander, binary.BigEndian, &v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
