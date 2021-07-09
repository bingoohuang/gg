package osx

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// GetGroupID the current user's group ID.
func GetGroupID() int64 {
	if out, err := exec.Command("id", "-g").Output(); err == nil {
		v := strings.TrimSpace(string(out))
		if g, e := strconv.ParseInt(v, 10, 32); e == nil {
			return g
		}
	}
	return -1
}

// IsRootGroup tells the current user's group is zero or not.
func IsRootGroup() bool {
	return GetGroupID() == 0
}

// CreateLogDir creates a log path.
// If logDir is empty, /var/log/{appName} or ~/logs/{appName} will used as logDir.
// If appName is empty, os.Args[0]'s base will be use as appName.
func CreateLogDir(logDir, appName string) string {
	if appName == "" {
		appName = filepath.Base(os.Args[0])
	}

	if logDir == "" {
		if IsRootGroup() {
			logDir = filepath.Join("/var/log/", appName)
		} else {
			logDir = filepath.Join("~/logs/" + appName)
		}
	} else {
		if stat, err := os.Stat(logDir); err == nil && !stat.IsDir() {
			logDir = filepath.Dir(logDir)
		}
	}

	logDir = ExpandHome(logDir)
	if stat, err := os.Stat(logDir); err == nil && stat.IsDir() {
		return logDir
	}

	syscall.Umask(0)
	if err := os.MkdirAll(logDir, os.ModeSticky|os.ModePerm); err != nil {
		panic(err)
	}

	return logDir
}
