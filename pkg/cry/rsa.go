package cry

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// from https://gist.github.com/miguelmota/3ea9286bd1d3c2a985b67cac4ba2130a
// and https://gist.github.com/sohamkamani/08377222d5e3e6bc130827f83b0c073e
// and https://asecuritysite.com/encryption/gorsa

// GenerateKeyPair generates a new key pair.
func GenerateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey) {
	key, _ := rsa.GenerateKey(rand.Reader, bits)
	return key, &key.PublicKey
}

// ExportPrivateKey exports private key to bytes.
func ExportPrivateKey(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

// ParsePrivateKey parses private key bytes to *rsa.PrivateKey.
func ParsePrivateKey(key []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// ExportPublicKey exports public key to bytes.
func ExportPublicKey(pub *rsa.PublicKey) []byte {
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes})
}

// ParsePublicKey parses public key bytes to *rsa.PublicKey.
func ParsePublicKey(key []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	if p, ok := pub.(*rsa.PublicKey); ok {
		return p, nil
	}

	return nil, errors.New("key type is not RSA")
}

// EncryptWithPublicKey encrypts data with public key.
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(sha512.New(), rand.Reader, pub, msg, nil)
}

// DecryptWithPrivateKey decrypts data with private key.
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	return rsa.DecryptOAEP(sha512.New(), rand.Reader, priv, ciphertext, nil)
}
