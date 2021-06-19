package flagparse

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"os"
)

func LoadConfFile(confFile, defaultConfFile string, app interface{}) error {
	if confFile == "" {
		if s, err := os.Stat(defaultConfFile); err != nil || s.IsDir() {
			return nil // not exists
		}
		confFile = defaultConfFile
	}

	data, err := os.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("read conf file %s error: %q", confFile, err)
	}

	if err := yaml.Unmarshal(data, app); err != nil {
		return fmt.Errorf("decode conf file %s error:%q", confFile, err)
	}

	return nil
}
