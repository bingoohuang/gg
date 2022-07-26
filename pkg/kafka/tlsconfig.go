package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

type TlsConfig struct {
	CaFile, CertFile, KeyFile string
	InsecureSkipVerify        bool
}

func (tc TlsConfig) Create() *tls.Config {
	if tc.CertFile == "" || tc.KeyFile == "" || tc.CaFile == "" {
		// will be nil by default if nothing is provided
		return nil
	}

	cert, err := tls.LoadX509KeyPair(tc.CertFile, tc.KeyFile)
	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile(tc.CaFile)
	if err != nil {
		log.Fatal(err)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            pool,
		InsecureSkipVerify: tc.InsecureSkipVerify,
	}
}
