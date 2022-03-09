package netx

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/mail"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/gg/pkg/timex"

	"golang.org/x/net/idna"
)

var (
	userAndHostname string

	RootCaName  = "root.pem"
	RootKeyName = "root.key"
)

type MkCert struct {
	Pkcs12, Ecdsa, Client, Silent bool
	KeyFile, CertFile, P12File    string
	CaRoot, CsrPath               string
	RootDuration, CertDuration    time.Duration

	caCert *x509.Certificate
	caKey  crypto.PrivateKey
}

func (m *MkCert) Run(hosts ...string) error {
	if m.CaRoot == "" {
		m.CaRoot = getCaRoot()
	}
	if m.CaRoot == "" {
		return fmt.Errorf("failed to find the default CA location, set env CAROOT")
	}
	if err := os.MkdirAll(m.CaRoot, 0o755); err != nil {
		return fmt.Errorf("create the CaRoot failed: %w", err)
	}

	if err := m.loadCaRoot(); err != nil {
		return err
	}

	if m.CsrPath != "" {
		return m.makeCertFromCSR()
	}

	if len(hosts) == 0 {
		return fmt.Errorf("at least one IP/host/email should be specified")
	}

	hostnameRegexp := regexp.MustCompile(`(?i)^(\*\.)?[0-9a-z_-]([0-9a-z._-]*[0-9a-z_-])?$`)
	for i, name := range hosts {
		if ip := net.ParseIP(name); ip != nil {
			continue
		}
		if email, err := mail.ParseAddress(name); err == nil && email.Address == name {
			continue
		}
		if uriName, err := url.Parse(name); err == nil && uriName.Scheme != "" && uriName.Host != "" {
			continue
		}
		punycode, err := idna.ToASCII(name)
		if err != nil {
			return fmt.Errorf("%q is not a valid hostname, IP, URL or email, failed: %w", name, err)
		}
		hosts[i] = punycode
		if !hostnameRegexp.MatchString(punycode) {
			return fmt.Errorf("%q is not a valid hostname, IP, URL or email", name)
		}
	}

	return m.makeCert(hosts)
}

func getCaRoot() string {
	if env := os.Getenv("CAROOT"); env != "" {
		return env
	}

	var dir string
	switch {
	case runtime.GOOS == "windows":
		dir = os.Getenv("LocalAppData")
	case os.Getenv("XDG_DATA_HOME") != "":
		dir = os.Getenv("XDG_DATA_HOME")
	case runtime.GOOS == "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return ""
		}
		dir = filepath.Join(dir, "Library", "Application Support")
	default: // Unix
		dir = os.Getenv("HOME")
		if dir == "" {
			return ""
		}
		dir = filepath.Join(dir, ".local", "share")
	}
	return filepath.Join(dir, ".cert")
}

func init() {
	u, err := user.Current()
	if err == nil {
		userAndHostname = u.Username + "@"
	}
	if h, err := os.Hostname(); err == nil {
		userAndHostname += h
	}
	if err == nil && u.Name != "" && u.Name != u.Username {
		userAndHostname += " (" + u.Name + ")"
	}
}

func (m *MkCert) makeCert(hosts []string) error {
	if m.caKey == nil {
		return fmt.Errorf("CA key (%s) is missing", RootKeyName)
	}

	priv, err := m.generateKey(false)
	if err != nil {
		return fmt.Errorf("generate certificate key failed: %w", err)
	}

	pub := priv.(crypto.Signer).Public()

	// Certificates last for 2 years and 3 months, which is always less than
	// 825 days, the limit that macOS/iOS apply to all certificates,
	// including custom roots. See https://support.apple.com/en-us/HT210176.
	if m.CertDuration == 0 {
		m.CertDuration, _ = timex.ParseDuration("2y3M")
	}

	serialNumber, err := randomSerialNumber()
	if err != nil {
		return err
	}

	start := time.Now().UTC()
	expiration := start.Add(m.CertDuration)

	c := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"mkcert dev certificate"},
			OrganizationalUnit: []string{userAndHostname},
		},

		NotBefore: start, NotAfter: expiration,
		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			c.IPAddresses = append(c.IPAddresses, ip)
		} else if email, err := mail.ParseAddress(h); err == nil && email.Address == h {
			c.EmailAddresses = append(c.EmailAddresses, h)
		} else if uriName, err := url.Parse(h); err == nil && uriName.Scheme != "" && uriName.Host != "" {
			c.URIs = append(c.URIs, uriName)
		} else {
			c.DNSNames = append(c.DNSNames, h)
		}
	}

	if m.Client {
		c.ExtKeyUsage = append(c.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}
	if len(c.IPAddresses) > 0 || len(c.DNSNames) > 0 || len(c.URIs) > 0 {
		c.ExtKeyUsage = append(c.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}
	if len(c.EmailAddresses) > 0 {
		c.ExtKeyUsage = append(c.ExtKeyUsage, x509.ExtKeyUsageEmailProtection)
	}

	// IIS (the main target of PKCS #12 files), only shows the deprecated
	// Common Name in the UI. See issue #115.
	if m.Pkcs12 {
		c.Subject.CommonName = hosts[0]
	}

	cert, err := x509.CreateCertificate(rand.Reader, c, m.caCert, pub, m.caKey)
	if err != nil {
		return fmt.Errorf("generate certificat failed: %w", err)
	}

	m.CertFile, m.KeyFile, m.P12File = m.fileNames(hosts)

	changeIt := "changeit"
	if m.Pkcs12 {
		domainCert, _ := x509.ParseCertificate(cert)
		pfxData, err := pkcs12.Encode(rand.Reader, priv, domainCert, []*x509.Certificate{m.caCert}, changeIt)
		if err != nil {
			return fmt.Errorf("generate PKCS#12 failed: %w", err)
		}

		if err := ioutil.WriteFile(m.P12File, pfxData, 0o644); err != nil {
			return fmt.Errorf("save PKCS#12 failed: %w", err)
		}
	} else {
		certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
		privDER, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			return fmt.Errorf("encode certificate key failed: %w", err)
		}
		privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})

		if m.CertFile == m.KeyFile {
			if err := ioutil.WriteFile(m.KeyFile, append(certPEM, privPEM...), 0o600); err != nil {
				return fmt.Errorf("save certificate and key failed: %w", err)
			}
		} else {
			if err := ioutil.WriteFile(m.CertFile, certPEM, 0o644); err != nil {
				return fmt.Errorf("save certificate failed: %w", err)
			}
			if err := ioutil.WriteFile(m.KeyFile, privPEM, 0o600); err != nil {
				return fmt.Errorf("save certificate key failed: %w", err)
			}
		}
	}

	m.printHosts(hosts)

	if !m.Silent {
		if m.Pkcs12 {
			log.Printf("The PKCS#12 bundle is at %q ✅", m.P12File)
			log.Printf("The legacy PKCS#12 encryption password is the often hardcoded default %q ℹ️", changeIt)
		} else {
			if m.CertFile == m.KeyFile {
				log.Printf("The certificate and key are at %q ✅", m.CertFile)
			} else {
				log.Printf("The certificate is at %q and the key at %q ✅", m.CertFile, m.KeyFile)
			}
		}

		log.Printf("The certificate will expire at %s 🗓", expiration.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func (m *MkCert) printHosts(hosts []string) {
	if m.Silent {
		return
	}

	secondLvlWildcardRegexp := regexp.MustCompile(`(?i)^\*\.[0-9a-z_-]+$`)
	for _, h := range hosts {
		log.Printf("Created a new certificate valid for the name - %q 📜", h)
		if secondLvlWildcardRegexp.MatchString(h) {
			log.Printf("Warning: many browsers don't support second-level wildcards like %q ⚠️", h)
		}
	}

	for _, h := range hosts {
		if strings.HasPrefix(h, "*.") {
			log.Printf("Reminder: X.509 wildcards only go one level deep, so this won't match a.b.%s ℹ️", h[2:])
			break
		}
	}
}

func (m *MkCert) generateKey(rootCA bool) (crypto.PrivateKey, error) {
	if m.Ecdsa {
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	if rootCA {
		return rsa.GenerateKey(rand.Reader, 3072)
	}

	return rsa.GenerateKey(rand.Reader, 2048)
}

func (m *MkCert) fileNames(hosts []string) (certFile, keyFile, p12File string) {
	defaultName := strings.Replace(hosts[0], ":", "_", -1)
	defaultName = strings.Replace(defaultName, "*", "_wildcard", -1)
	if len(hosts) > 1 {
		defaultName += "+" + strconv.Itoa(len(hosts)-1)
	}
	if m.Client {
		defaultName += "-client"
	}

	certFile = filepath.Join(m.CaRoot, defaultName+".pem")
	if m.CertFile != "" {
		certFile = m.CertFile
	}
	keyFile = filepath.Join(m.CaRoot, defaultName+".key")
	if m.KeyFile != "" {
		keyFile = m.KeyFile
	}
	p12File = filepath.Join(m.CaRoot, defaultName+".p12")
	if m.P12File != "" {
		p12File = m.P12File
	}

	return
}

func randomSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("generate serial number failed: %w", err)
	}
	return serialNumber, nil
}

func (m *MkCert) makeCertFromCSR() error {
	if m.caKey == nil {
		return fmt.Errorf("can't create new certificates because the CA key (rootCA-key.pem) is missing")
	}

	csrPEMBytes, err := ioutil.ReadFile(m.CsrPath)
	if err != nil {
		return fmt.Errorf("read the CSR %s failed: %w", m.CsrPath, err)
	}

	csrPEM, _ := pem.Decode(csrPEMBytes)
	if csrPEM == nil {
		return fmt.Errorf("read the CSR failed: unexpected content")
	}

	if csrPEM.Type != "CERTIFICATE REQUEST" && csrPEM.Type != "NEW CERTIFICATE REQUEST" {
		return fmt.Errorf("read the CSR failed, expect CERTIFICATE REQUEST, got %s", csrPEM.Type)
	}
	csr, err := x509.ParseCertificateRequest(csrPEM.Bytes)
	if err != nil {
		return fmt.Errorf("parse the CSR %s failed: %w", m.CsrPath, err)
	}
	if err := csr.CheckSignature(); err != nil {
		return fmt.Errorf("check CSR %s signature failed: %w", m.CsrPath, err)
	}

	if m.CertDuration == 0 {
		m.CertDuration, _ = timex.ParseDuration("2y3M")
	}
	serialNumber, err := randomSerialNumber()
	if err != nil {
		return err
	}

	start := time.Now().UTC()
	expiration := start.Add(m.CertDuration)

	tpl := &x509.Certificate{
		SerialNumber:    serialNumber,
		Subject:         csr.Subject,
		ExtraExtensions: csr.Extensions, // includes requested SANs, KUs and EKUs

		NotBefore: start, NotAfter: expiration,

		// If the CSR does not request a SAN extension, fix it up for them as
		// the Common Name field does not work in modern browsers. Otherwise,
		// this will get overridden.
		DNSNames: []string{csr.Subject.CommonName},

		// Likewise, if the CSR does not set KUs and EKUs, fix it up as Apple
		// platforms require serverAuth for TLS.
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	if m.Client {
		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}
	if len(csr.EmailAddresses) > 0 {
		tpl.ExtKeyUsage = append(tpl.ExtKeyUsage, x509.ExtKeyUsageEmailProtection)
	}

	cert, err := x509.CreateCertificate(rand.Reader, tpl, m.caCert, csr.PublicKey, m.caKey)
	if err != nil {
		return fmt.Errorf("generate certificate failed: %w", err)
	}

	var hosts []string
	hosts = append(hosts, csr.DNSNames...)
	hosts = append(hosts, csr.EmailAddresses...)
	for _, ip := range csr.IPAddresses {
		hosts = append(hosts, ip.String())
	}
	for _, uri := range csr.URIs {
		hosts = append(hosts, uri.String())
	}
	certFile, _, _ := m.fileNames(hosts)

	pemMem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	if err := ioutil.WriteFile(certFile, pemMem, 0o644); err != nil {
		return fmt.Errorf("save certificate failed: %w", err)
	}

	m.printHosts(hosts)

	if !m.Silent {
		log.Printf("The certificate is at %q ✅, it will expire at %s 🗓",
			certFile, expiration.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// loadCaRoot will load or create the CA at CaRoot.
func (m *MkCert) loadCaRoot() error {
	if !PathExists(filepath.Join(m.CaRoot, RootCaName)) {
		if err := m.newCA(); err != nil {
			return err
		}
	}

	certPEMBlock, err := ioutil.ReadFile(filepath.Join(m.CaRoot, RootCaName))
	if err != nil {
		return fmt.Errorf("read the CA certificat failed: %w", err)
	}
	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil || certDERBlock.Type != "CERTIFICATE" {
		return fmt.Errorf("read the CA certificat failed: unexpected content")
	}
	m.caCert, err = x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parse the CA certificat failed: %w", err)
	}

	if !PathExists(filepath.Join(m.CaRoot, RootKeyName)) {
		return nil // keyless mode, where only -install works
	}

	keyPEMBlock, err := ioutil.ReadFile(filepath.Join(m.CaRoot, RootKeyName))
	if err != nil {
		return fmt.Errorf("read the CA key failed: %w", err)
	}
	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	if keyDERBlock == nil || keyDERBlock.Type != "PRIVATE KEY" {
		return fmt.Errorf("read the CA key failed: unexpected content")
	}
	m.caKey, err = x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parse the CA key failed: %w", err)
	}

	return nil
}

func (m *MkCert) newCA() error {
	priv, err := m.generateKey(true)
	if err != nil {
		return fmt.Errorf("generate the CA key failed: %w", err)
	}

	pub := priv.(crypto.Signer).Public()

	spkiASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return fmt.Errorf("encode public key failed: %w", err)
	}

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	if err != nil {
		return fmt.Errorf("decode public key failed: %w", err)
	}

	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)
	serialNumber, err := randomSerialNumber()
	if err != nil {
		return err
	}

	if m.RootDuration == 0 {
		m.RootDuration, _ = timex.ParseDuration("10y")
	}

	start := time.Now().UTC()
	expiration := start.Add(m.RootDuration)

	tpl := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"mkcert development CA"},
			OrganizationalUnit: []string{userAndHostname},

			// The CommonName is required by iOS to show the certificate in the
			// "Certificate Trust Settings" menu.
			// https://github.com/FiloSottile/mkcert/issues/47
			CommonName: "mkcert " + userAndHostname,
		},
		SubjectKeyId: skid[:],

		NotBefore: start, NotAfter: expiration,

		KeyUsage: x509.KeyUsageCertSign,

		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv)
	if err != nil {
		return fmt.Errorf("generate CA certificate failed: %w", err)
	}

	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("encode CA key failed: %w", err)
	}
	if err = ioutil.WriteFile(filepath.Join(m.CaRoot, RootKeyName), pem.EncodeToMemory(
		&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}), 0o400); err != nil {
		return fmt.Errorf("save CA key failed: %w", err)
	}
	if err = ioutil.WriteFile(filepath.Join(m.CaRoot, RootCaName), pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: cert}), 0o644); err != nil {
		return fmt.Errorf("save CA certificate failed: %w", err)
	}

	if !m.Silent {
		log.Printf("Created a new local CA ✅")
	}
	return nil
}
