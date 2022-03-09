package netx

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/bingoohuang/gg/pkg/ss"
)

type File struct {
	Cert string
	Key  string
}

// LoadCerts loads an existing certificate and key or creates new.
func LoadCerts(caRoot string) *File {
	certFile, err := ParseCerts(caRoot)
	if err != nil {
		log.Fatalf("parse certs failed: %v", err)
	} else if certFile != nil {
		log.Printf("cert found %+v", *certFile)
	} else {
		if caRoot == "" {
			caRoot = ".cert"
		}
		mk := MkCert{CaRoot: caRoot}
		if err := mk.Run("localhost"); err != nil {
			log.Fatalf("mkcert failed: %v", err)
		}
		certFile = &File{
			Cert: mk.CertFile,
			Key:  mk.KeyFile,
		}
	}

	return certFile
}

// ParseCerts tries to parse the certificate and key in the certPath.
func ParseCerts(certPath string) (*File, error) {
	if certPath == "" {
		return nil, nil
	}

	stat, err := os.Stat(certPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat %s failed: %v", certPath, err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a dir", certPath)
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

	if len(keyFiles) == 0 && len(certFiles) == 0 {
		return nil, nil
	}

	if len(keyFiles) == 1 && len(certFiles) == 1 {
		return &File{
			Cert: certFiles[0],
			Key:  keyFiles[0],
		}, nil
	}

	// filter root
	var filterKeyFiles, filterCertFiles []string
	for _, k := range keyFiles {
		if !ss.ContainsFold(k, "root") {
			filterKeyFiles = append(filterKeyFiles, k)
		}
	}

	for _, k := range certFiles {
		if !ss.ContainsFold(k, "root") {
			filterCertFiles = append(filterCertFiles, k)
		}
	}

	if len(filterKeyFiles) == 1 && len(filterCertFiles) == 1 {
		return &File{
			Cert: certFiles[0],
			Key:  keyFiles[0],
		}, nil
	}

	return nil, fmt.Errorf("multiple keyFiles %v and certFiles %v found", keyFiles, certFiles)
}
