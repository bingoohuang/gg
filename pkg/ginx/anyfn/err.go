package anyfn

// AdapterError defines the error generated in the processing of adaptor.
type AdapterError struct {
	Err     error
	Context string
}

// Error returns the error message.
func (e *AdapterError) Error() string { return e.Context + " " + e.Err.Error() }

// IsAdapterError tells if err is an AdaptorError or not.
func IsAdapterError(err error) bool { _, ok := err.(*AdapterError); return ok }
