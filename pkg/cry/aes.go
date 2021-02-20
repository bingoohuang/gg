package cry

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// AESOpt contains all aes session option
type AESOpt struct {
	aesGCM cipher.AEAD
}

// NewAESOpt is function to create new configuration of aes algorithm option
// the secret must be hexa a-f & 0-9
func NewAESOpt(secret string) (*AESOpt, error) {
	if len(secret) != 64 {
		return nil, errors.New("Secret must be 64 character")
	}
	key, err := hex.DecodeString(secret)
	if err != nil {
		return nil, errors.Wrap(err, "NewAESOpt.hex.DecodeString")
	}

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "NewAESOpt.aes.NewCipher")
	}

	// AES本身有四种可能的模式，但是常用的只有两种：CBC和GCM。
	// CBC模式就是一个一个块的计算每个块的编码结果，下一个块输入上一个块的计算结果，第一个块输入的是IV（初始向量）。如此所有的数据块形成一条链，逐个计算得到了最终的加密数据。
	// GCM来自于AES的CTR模式，CTR是指计数器模式。GCM是利用GMAC（基于伽罗华域的MAC）和AES的CTR模式的组合。GMAC比普通的MAC算法快（毕竟冠以伽罗华之名），
	// GCM模式与CBC的一个最大的区别是GCM模式不再把上一个数据块的计算结果输入到下一个数据块的计算，而是在分好的数据块中任意位置开始计算，
	// 由一个计数器和一个不变的IV值（nounce）来控制每一次计算的随机性。由于下一次的计算并不依赖于上一次的结果，所以GCM模式可以实现大规模的并行化，并且Intel还专门推出了clmul指令用于加速GCM的运算速度，可见其应用之广。
	// 推荐先后顺序：GCM > CTR > CBC http://nanhuacoder.top/2018/01/03/iOS-UseAES
	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "NewAESOpt.cipher.NewGCM")
	}

	return &AESOpt{
		aesGCM: aesGCM,
	}, nil
}

// Encrypt is function to encrypt data using AES algorithm
func (aesOpt *AESOpt) Encrypt(plainText []byte) (string, error) {

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesOpt.aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "encryptAES.io.ReadFull")
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesOpt.aesGCM.Seal(nonce, nonce, plainText, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

// Decrypt is function to decypt data using AES algorithm
func (aesOpt *AESOpt) Decrypt(chiperText []byte) (string, error) {

	enc, _ := hex.DecodeString(string(chiperText))

	//Get the nonce size
	nonceSize := aesOpt.aesGCM.NonceSize()
	if len(enc) < nonceSize {
		return "", errors.New("The data can't be decrypted")
	}
	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plainText, err := aesOpt.aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.Wrap(err, "decryptAES.aesGCM.Open")
	}

	return fmt.Sprintf("%s", plainText), nil
}
