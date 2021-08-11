package ss

import "strings"

// Split1 将s按分隔符sep分成x份，取第1份
func Split1(s string, options ...SplitOption) (s0 string) {
	s0, _, _, _, _ = Split5(s, options...)
	return
}

// Split2 将s按分隔符sep分成x份，取第1、2份
func Split2(s string, options ...SplitOption) (s0, s1 string) {
	s0, s1, _, _, _ = Split5(s, options...)
	return
}

// Split3 将s按分隔符sep分成x份，取第1、2、3份
func Split3(s string, options ...SplitOption) (s0, s1, s2 string) {
	s0, s1, s2, _, _ = Split5(s, options...)
	return
}

// Split4 将s按分隔符sep分成x份，取第x份，取第1、2、3、4份
func Split4(s string, options ...SplitOption) (s0, s1, s2, s3 string) {
	s0, s1, s2, s3, _ = Split5(s, options...)
	return
}

// Split5 将s按分隔符sep分成x份，取第x份，取第1、2、3、4、5份
func Split5(s string, options ...SplitOption) (s0, s1, s2, s3, s4 string) {
	parts := Split(s, options...)
	l := len(parts)

	if l > 0 {
		s0 = strings.TrimSpace(parts[0])
	}
	if l > 1 {
		s1 = strings.TrimSpace(parts[1])
	}
	if l > 2 {
		s2 = strings.TrimSpace(parts[2])
	}
	if l > 3 {
		s3 = strings.TrimSpace(parts[3])
	}
	if l > 4 {
		s4 = strings.TrimSpace(parts[4])
	}
	return
}

// SplitToMap 将字符串s分割成map,其中key和value之间的间隔符是kvSep, kv和kv之间的分隔符是kkSep
func SplitToMap(s string, kvSep, kkSep string) map[string]string {
	var m map[string]string

	ss := strings.Split(s, kkSep)
	m = make(map[string]string)

	for _, pair := range ss {
		p := strings.TrimSpace(pair)
		if p == "" {
			continue
		}

		k, v := Split2(p, WithSeps(kvSep), WithIgnoreEmpty(false))
		m[k] = v
	}

	return m
}
