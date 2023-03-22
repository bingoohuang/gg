package cry_test

import (
	"fmt"
	"log"

	"github.com/bingoohuang/gg/pkg/cry"
)

func ExampleNewAESOpt() {
	// sample plain text
	sampleText := "Halo this is encrypted text!!!"

	// it's random string must be hexa  a-f & 0-9
	const key = "fa89277fb1e1c344709190deeac4465c2b28396423c8534a90c86322d0ec9dcf"
	// define AES option
	aesOpt, err := cry.NewAESOpt(key)
	if err != nil {
		log.Println("ERR", err)
		return
	}

	// Encrypt text using AES algorithm
	cipherText, err := aesOpt.Encrypt([]byte(sampleText))
	if err != nil {
		log.Println("ERR", err)
		return
	}
	// 每次加密结果不一样（因为加入了随机的nonce)
	// fmt.Println("Encrypted data", string(cipherText))
	// Encrypted data 28509ae15364b3994846eeb4079122825285ac7863f94fab43aacb43b5a81e2f93c58743c5d089ae4fefe468b1e240ac88cb50e9316a2f618d55
	// Encrypted data b91a0019778033730704da5881a39c955f58c878f6a7938da107a2cbb9d752100af56024463d901c75b1ea6cb63089bf88e8b758b9786de9150e

	// Decrypt text using AES algorithm
	plainText, err := aesOpt.Decrypt([]byte(cipherText))
	if err != nil {
		log.Println("ERR", err)
		return
	}
	fmt.Println("Decrypted data", string(plainText))

	// Output:
	// Decrypted data Halo this is encrypted text!!!
}

func ExampleNewDESOpt() {
	// sample plain text
	sampleText := "Halo this is encrypted text!!!"
	// it's character 24 bit
	const key = "123456781234567812345678"
	// define DES option
	desOpt, err := cry.NewDESOpt(key)
	if err != nil {
		log.Println("ERR", err)
		return
	}

	// Encrypt text using DES algorithm
	cipherText, err := desOpt.Encrypt([]byte(sampleText))
	if err != nil {
		log.Println("ERR", err)
		return
	}
	// fmt.Println("Encrypted data:", cipherText)
	// 每次都得到一样的数据
	// Encrypted data: k1Uoi4OsCMBSCeVxdBmwfVuO2PxndJZSfCsXIULB7F0=

	// Decrypt text using DES algorithm
	plainText, err := desOpt.Decrypt([]byte(cipherText))
	if err != nil {
		log.Println("ERR", err)
		return
	}
	fmt.Println("Decrypted data:", plainText)
	// Output:
	// Decrypted data: Halo this is encrypted text!!!
}

func ExampleNewRC4Opt() {
	// sample plain text
	sampleText := "Halo this is encrypted text!!!"
	// it's character 24 bit
	const key = "123456781234567812345678"
	// define RC4 option
	rc4Opt, err := cry.NewRC4Opt(key)
	if err != nil {
		log.Println("ERR", err)
		return
	}

	// Encrypt text using RC4 algorithm
	cipherText, err := rc4Opt.Encrypt([]byte(sampleText))
	if err != nil {
		log.Println("ERR", err)
		return
	}
	// fmt.Println("Encrypted data:", cipherText)
	// 每次都得到一样的数据
	// Encrypted data: f39255bb29c5b6ce831363bc865866c600e7ed3dac0c5dc13c63a196788c

	// Decrypt text using RC4 algorithm
	plainText, err := rc4Opt.Decrypt([]byte(cipherText))
	if err != nil {
		log.Println("ERR", err)
		return
	}
	fmt.Println("Decrypted data:", plainText)
	// Output:
	// Decrypted data: Halo this is encrypted text!!!
}
