package human_test

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/human"
)

func ExampleBytes() {
	fmt.Printf("That file is %s.\n", human.Bytes(82854982))
	fmt.Printf("That file is %s.\n", human.IBytes(82854982))
	fmt.Printf("That file is %s.\n", human.ByteCount(82854982))
	fmt.Printf("That file is %s.\n", human.IByteCount(82854982))

	// Output:
	// That file is 82.9MB.
	// That file is 79MiB.
	// That file is 82.9MB.
	// That file is 79MiB.
}
