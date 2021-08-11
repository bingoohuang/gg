package mapp

// GetStringOr returns the value associated to the key,
// or return defValue when value does not exit, or it is empty.
func GetStringOr(m map[string]string, key, defValue string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}

	return defValue
}
