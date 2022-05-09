package sigx

import (
	"bytes"
	"context"
	"github.com/bingoohuang/gg/pkg/iox"
	"github.com/bingoohuang/gg/pkg/osx"
	"github.com/bingoohuang/gg/pkg/profile"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"
)

// RegisterSignals registers signal handlers.
func RegisterSignals(c context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	if c == nil {
		c = context.Background()
	}
	cc, cancel := context.WithCancel(c)
	sig := make(chan os.Signal, 1)
	if len(signals) == 0 {
		// syscall.SIGINT: ctl + c, syscall.SIGTERM: kill pid
		signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	signal.Notify(sig, signals...)
	go func() {
		<-sig
		cancel()
	}()

	return cc, cancel
}

func RegisterSignalCallback(f func(), signals ...os.Signal) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, signals...)
	go func() {
		for range sig {
			f()
		}
	}()
}

func RegisterSignalProfile(signals ...os.Signal) {
	if len(signals) == 0 {
		signals = defaultSignals
	}

	RegisterSignalCallback(func() {
		val, ok := ReadCmdOK("jj.cpu", false)
		if ok {
			collectCpuProfile()
			if val = bytes.TrimSpace(val); len(val) > 0 {
				if duration, err := time.ParseDuration(string(val)); err != nil {
					log.Printf("ignore duration %s in jj.cpu, parse failed: %v", val, err)
				} else if duration > 0 {
					log.Printf("after %s, cpu.profile will be generated ", val)
					go func() {
						time.Sleep(duration)
						collectCpuProfile()
					}()
				}
			}
		}

		if HasCmd("jj.mem", true) {
			if err := CollectMemProfile("mem.profile"); err != nil {
				log.Printf("failed to collect profile: %v", err)
			}
		}
		if v := ReadCmd("jj.profile"); len(v) > 0 {
			go profile.Start(profile.Specs(string(v)))
		}
	}, signals...)
}

func collectCpuProfile() {
	if completed, err := CollectCpuProfile("cpu.profile"); err != nil {
		log.Printf("failed to collect profile: %v", err)
	} else if completed {
		osx.Remove("jj.cpu")
	}
}

var cpuProfileFile *os.File

func HasCmd(f string, remove bool) bool {
	s, err := os.Stat(f)
	if err == nil && !s.IsDir() {
		if remove {
			osx.Remove(f)
		}
		return true
	}

	return false
}

func ReadCmdOK(f string, remove bool) ([]byte, bool) {
	s, err := os.Stat(f)
	if err == nil && !s.IsDir() {
		data, _ := ioutil.ReadFile(f)
		if remove {
			osx.Remove(f)
		}
		return data, true
	}

	return nil, false
}

func ReadCmd(f string) []byte {
	s, err := os.Stat(f)
	if err == nil && !s.IsDir() {
		data, _ := ioutil.ReadFile(f)
		osx.Remove(f)
		return data
	}

	return nil
}

func CollectCpuProfile(cpuProfile string) (bool, error) {
	if cpuProfile == "" {
		return false, nil
	}

	if cpuProfileFile != nil {
		pprof.StopCPUProfile()
		iox.Close(cpuProfileFile)

		log.Printf("%s collected", cpuProfileFile.Name())
		cpuProfileFile = nil
		return true, nil
	}

	f, err := os.Create(cpuProfile)
	if err != nil {
		return false, err
	}
	cpuProfileFile = f

	if err := pprof.StartCPUProfile(f); err != nil {
		return false, err
	}

	log.Printf("%s started", cpuProfile)
	return false, nil
}

func CollectMemProfile(memProfile string) error {
	if memProfile == "" {
		return nil
	}

	f, err := os.Create(memProfile)
	if err != nil {
		return err
	}
	defer f.Close()

	return pprof.WriteHeapProfile(f)
}
