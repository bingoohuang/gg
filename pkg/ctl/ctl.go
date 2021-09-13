package ctl

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/bingoohuang/gg/pkg/v"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	_ "embed"
)

type Config struct {
	Initing      bool
	PrintVersion bool
	InitFiles    *embed.FS
}

// ProcessInit generates ctl and conf.yml files for -init argument.
func (c Config) ProcessInit() {
	if c.Initing {
		if err := initCtl(); err != nil {
			fmt.Printf(err.Error())
		}
		if err := c.initFiles(); err != nil {
			fmt.Printf(err.Error())
		}
	}

	if c.PrintVersion {
		fmt.Print(v.Version())
	}

	if c.Initing || c.PrintVersion {
		os.Exit(0)
	}
}

func (c Config) initFiles() error {
	firstDir := false
	return fs.WalkDir(c.InitFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		if d.IsDir() {
			if firstDir {
				return fs.SkipDir
			}

			firstDir = true
			return nil
		}

		if _, err := os.Stat(d.Name()); err == nil {
			fmt.Println(d.Name(), "already exists, ignored!")
			return nil
		}

		// 0644->即用户具有读写权限，组用户和其它用户具有只读权限；
		data, err := c.InitFiles.ReadFile(path)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(d.Name(), data, 0o644); err != nil {
			return err
		}

		fmt.Println(d.Name(), "created!")
		return nil
	})
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
