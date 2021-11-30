package sigx

import (
	"context"
	"github.com/bingoohuang/gg/pkg/profile"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
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
	RegisterSignalCallback(func() {
		if HasCmd("jj.cpu") {
			if err := CollectCpuProfile("cpu.profile"); err != nil {
				log.Printf("failed to collect profile: %v", err)
			}
		}
		if HasCmd("jj.mem") {
			if err := CollectMemProfile("mem.profile"); err != nil {
				log.Printf("failed to collect profile: %v", err)
			}
		}
		if v := ReadCmd("jj.profile"); len(v) > 0 {
			go profile.Start(profile.Specs(string(v)))
		}
	}, signals...)
}

var cpuProfileFile *os.File

func HasCmd(f string) bool {
	s, err := os.Stat(f)
	if err == nil && !s.IsDir() {
		os.Remove(f)
		return true
	}

	return false
}

func ReadCmd(f string) []byte {
	s, err := os.Stat(f)
	if err == nil && !s.IsDir() {
		data, _ := ioutil.ReadFile(f)
		os.Remove(f)
		return data
	}

	return nil
}

func CollectCpuProfile(cpuProfile string) error {
	if cpuProfile == "" {
		return nil
	}

	if cpuProfileFile != nil {
		pprof.StopCPUProfile()
		cpuProfileFile.Close()

		log.Printf("%s collected", cpuProfileFile.Name())
		cpuProfileFile = nil
		return nil
	}

	f, err := os.Create(cpuProfile)
	if err != nil {
		return err
	}
	cpuProfileFile = f

	if err := pprof.StartCPUProfile(f); err != nil {
		return err
	}

	log.Printf("%s started", cpuProfile)
	return nil
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
