package cry_test

import (
	cry "github.com/firdasafridi/gocrypt"
	"testing"
)

// https://golang.org/src/crypto/tls/generate_cert.go

func TestRSAPemTools(t *testing.T) {
	// Create the key pair.
	privateKey, publicKey := cry.GenerateKeyPair(2048)

	// Export the keys as PEM.
	privateKeyAsPEM := cry.ExportPrivateKey(privateKey)
	publicKeyAsPEM := cry.ExportPublicKey(publicKey)

	// Import the keys from PEM.
	privateKeyFromPEM, _ := cry.ParsePrivateKey(privateKeyAsPEM)
	publicKeyFromPEM, _ := cry.ParsePublicKey(publicKeyAsPEM)

	// Export the imported keys as PEM.
	privateKeyParsedAsPEM := cry.ExportPrivateKey(privateKeyFromPEM)
	publicKeyParsedAsPEM := cry.ExportPublicKey(publicKeyFromPEM)

	// Test if the exported/imported keys match the original keys.
	if string(privateKeyAsPEM) != string(privateKeyParsedAsPEM) ||
		string(publicKeyAsPEM) != string(publicKeyParsedAsPEM) {
		t.Error("Export and Import did not result in same Keys")
	}
}
