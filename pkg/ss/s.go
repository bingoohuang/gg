package ss

func Or(a, b string) string {
	if a == "" {
		return b
	}

	return a
}

func If(b bool, s1, s2 string) string {
	if b {
		return s1
	}

	return s2
}
