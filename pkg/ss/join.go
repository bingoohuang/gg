package ss

import "strings"

// JoinMap 将 s 合成位 string,其中key和value之间的间隔符是kvSep, kv和kv之间的分隔符是kkSep
func JoinMap(m map[string]string, kvSep, kkSep string) (s string) {
	items := make([]string, len(m))
	for k, v := range m {
		items = append(items, k+kvSep+v)
	}

	return strings.Join(items, kkSep)
}
