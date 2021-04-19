package xx

import (
	"bytes"
	"fmt"
	"log"
)

func ExampleGzip() {
	// define original data
	data := []byte(`MyzYrIyMLyNqwDSTBqSwM2D6KD9sA8S/d3Vyy6ldE+oRVdWyqNQrjTxQ6uG3XBOS0P4GGaIMJEPQ/gYZogwkQ+A0/gSU03fRJvdhIGQ1AMARVdWyqNQrjRFV1bKo1CuNEVXVsqjUK40RVdWyqNQrjRFV1bKo1CuNPmQF870PPsnSNeKI1U/MrOA0/gSU03fRb2A3OsnORNIruhCUYTIrOMTNU7JuGb5RSYJxa6PiMHdiRmFtXLNoY+GVmTD7aOV/K1yo4y0dR7Q=`)

	// compress data
	compressedData, compressedDataErr := Gzip(data)
	if compressedDataErr != nil {
		log.Fatal(compressedDataErr)
	}

	// uncompress data
	uncompressedData, uncompressedDataErr := Ungzip(compressedData)
	if uncompressedDataErr != nil {
		log.Fatal(uncompressedDataErr)
	}

	fmt.Println("check equal:", bytes.Equal(data, uncompressedData))
	// Output:
	// check equal: true
}
