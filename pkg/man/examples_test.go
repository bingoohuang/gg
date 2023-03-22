package man_test

import (
	"fmt"

	"github.com/bingoohuang/gg/pkg/man"
)

func ExampleBytes() {
	fmt.Printf("That file is %s.\n", man.Bytes(82854982))
	fmt.Printf("That file is %s.\n", man.IBytes(82854982))
	fmt.Printf("That file is %s.\n", man.ByteCount(82854982))
	fmt.Printf("That file is %s.\n", man.IByteCount(82854982))

	// Output:
	// That file is 82.9MB.
	// That file is 79MiB.
	// That file is 82.9MB.
	// That file is 79MiB.
}
