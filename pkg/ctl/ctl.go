package ctl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	_ "embed"
)

var COMMIT string

type Config struct {
	Initing      bool
	PrintVersion bool
	VersionInfo  string

	ConfTemplate []byte
	ConfFileName string
}

// ProcessInit generates ctl and conf.yml files for -init argument.
func (c Config) ProcessInit() {
	if c.Initing {
		if err := initCtl(); err != nil {
			fmt.Printf(err.Error())
		}
		if err := c.initConf(); err != nil {
			fmt.Printf(err.Error())
		}
	}

	if c.PrintVersion {
		fmt.Printf("Version: %s, COMMIT: %s\n", c.VersionInfo, COMMIT)
	}

	if c.Initing || c.PrintVersion {
		os.Exit(0)
	}
}

func (c Config) initConf() error {
	_, err := os.Stat(c.ConfFileName)
	if err == nil {
		fmt.Println(c.ConfFileName, "already exists, ignored!")
		return nil
	}

	// 0644->即用户具有读写权限，组用户和其它用户具有只读权限；
	if err := ioutil.WriteFile(c.ConfFileName, c.ConfTemplate, 0o644); err != nil {
		return err
	}

	fmt.Println(c.ConfFileName, "created!")
	return nil
}

//go:embed ctl
var ctlRaw string

func initCtl() error {
	_, err := os.Stat("ctl")
	if err == nil {
		fmt.Println("ctl already exists, ignored!")
		return nil
	}

	tpl, err := template.New("ctl").Parse(ctlRaw)
	if err != nil {
		return err
	}

	binArgs := argsExcludeInit()

	m := map[string]string{"BinName": os.Args[0], "BinArgs": strings.Join(binArgs, " ")}

	var content bytes.Buffer
	if err := tpl.Execute(&content, m); err != nil {
		return err
	}

	// 0755->即用户具有读/写/执行权限，组用户和其它用户具有读写权限；
	if err = ioutil.WriteFile("ctl", content.Bytes(), 0o755); err != nil {
		return err
	}

	fmt.Println("ctl created!")

	return nil
}

func argsExcludeInit() []string {
	binArgs := make([]string, 0, len(os.Args)-2)

	for i, arg := range os.Args {
		if i == 0 || strings.Index(arg, "-i") == 0 || strings.Index(arg, "--init") == 0 {
			continue
		}

		if strings.Index(arg, "-") != 0 {
			arg = strconv.Quote(arg)
		}

		binArgs = append(binArgs, arg)
	}

	return binArgs
}
