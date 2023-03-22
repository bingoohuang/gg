package sqx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryUrlEncodePass(t *testing.T) {
	assert.Equal(t, "postgres://SYSTEM:abc123%21%40%40%23@192.168.1.2:54321/mydb?sslmode=disable", tryUrlEncodePass("pgx", "postgres://SYSTEM:abc123!@@#@192.168.1.2:54321/mydb?sslmode=disable"))
	assert.Equal(t, "postgres://SYSTEM:abc123@192.168.1.2:54321/mydb?sslmode=disable", tryUrlEncodePass("pgx", "postgres://SYSTEM:abc123@192.168.1.2:54321/mydb?sslmode=disable"))
}
