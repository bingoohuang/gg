package gokv

type Store interface {
	// All list the keys in the store.
	All() (map[string]string, error)
	// Set stores the given value for the given key.
	Set(k, v string) error
	// Get retrieves the value for the given key.
	Get(k string) (v string, err error)
	// Del deletes the stored value for the given key.
	// Deleting a non-existing key-value pair does NOT lead to an error.
	Del(k string) error
}
