package iox

import (
	"io"
	"log"
)

// Close closes the io.Closer and log print if error occurs.
func Close(cc ...io.Closer) {
	for _, c := range cc {
		if c == nil {
			continue
		}

		if err := c.Close(); err != nil {
			log.Printf("close failed: %v", err)
		}
	}
}

// CloseAny closes any and log print if error occurs.
func CloseAny(cc ...interface{}) {
	for _, c := range cc {
		if c == nil {
			continue
		}

		if cl, ok := c.(io.Closer); ok {
			if err := cl.Close(); err != nil {
				log.Printf("close failed: %v", err)
			}
			continue
		}

		if cl, ok := c.(interface{ Close() }); ok {
			cl.Close()
			continue
		}
	}
}
