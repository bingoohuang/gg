package netx

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gg/pkg/ss"
)

type CertFiles struct {
	Cert string
	Key  string
}

// LoadCerts loads an existing certificate and key or creates new.
// CaRoot can be {dir}:{name}, like
// "" will default to .cert directory
// ":server" will find server.key and server.pem in .cert directory
// "." will default to xxx.key and xxx.pem in current directory
// ".:server" will find server.key and server.pem in . directory
func LoadCerts(caRoot string) *CertFiles {
	caRoot, certFile, err := ParseCerts(caRoot)
	if err != nil {
		log.Fatalf("parse certs failed: %v", err)
	} else if certFile != nil {
		log.Printf("cert found %+v", *certFile)
	} else {
		mk := MkCert{CaRoot: caRoot}
		if err := mk.Run("localhost"); err != nil {
			log.Fatalf("mkcert failed: %v", err)
		}
		certFile = &CertFiles{
			Cert: mk.CertFile,
			Key:  mk.KeyFile,
		}
	}

	return certFile
}

// ParseCerts tries to parse the certificate and key in the certPath.
func ParseCerts(certPath string) (string, *CertFiles, error) {
	specifiedName := ""
	lastCommaPos := strings.LastIndex(certPath, ":")
	if lastCommaPos >= 0 {
		specifiedName = certPath[lastCommaPos+1:]
		certPath = certPath[:lastCommaPos]
	}

	if certPath == "" {
		certPath = ".cert"
	}

	stat, err := os.Stat(certPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return certPath, nil, nil
		}
		return "", nil, fmt.Errorf("stat %s failed: %v", certPath, err)
	}
	if !stat.IsDir() {
		return "", nil, fmt.Errorf("%s is not a dir", certPath)
	}

	var keyFiles, certFiles []string
	_ = filepath.WalkDir(certPath, func(root string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if root == certPath {
				return nil
			}
			return fs.SkipDir
		}

		switch ext := path.Ext(info.Name()); {
		case ss.AnyOfFold(ext, ".key") || ss.ContainsFold(root, "-key."):
			keyFiles = append(keyFiles, root)
		case ss.AnyOfFold(ext, ".pem", ".crt"):
			certFiles = append(certFiles, root)
		}

		return nil
	})

	if len(certFiles) == 0 && len(keyFiles) == 0 {
		return certPath, nil, nil
	}

	if len(certFiles) == 1 && len(keyFiles) == 1 {
		return certPath, &CertFiles{Cert: certFiles[0], Key: keyFiles[0]}, nil
	}

	filter := func(input []string, sub string, included bool) (ret []string) {
		for _, k := range input {
			contains := ss.ContainsFold(k, sub)
			if included && contains || !included && !contains {
				ret = append(ret, k)
			}
		}
		return
	}

	if specifiedName != "" {
		specifiedCertFiles := filter(certFiles, specifiedName, true)
		specifiedKeyFiles := filter(keyFiles, specifiedName, true)
		if len(specifiedCertFiles) == 1 && len(specifiedKeyFiles) == 1 {
			return certPath, &CertFiles{Cert: specifiedCertFiles[0], Key: specifiedKeyFiles[0]}, nil
		}

	}
	filterCertFiles := filter(certFiles, "root", false)
	filterKeyFiles := filter(keyFiles, "root", false)
	if len(filterCertFiles) == 1 && len(filterKeyFiles) == 1 {
		return certPath, &CertFiles{Cert: filterCertFiles[0], Key: filterKeyFiles[0]}, nil
	}

	return "", nil, fmt.Errorf("multiple keyFiles %v and certFiles %v found", keyFiles, certFiles)
}
