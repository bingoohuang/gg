package profile

import (
	"fmt"
	"github.com/bingoohuang/gg/pkg/ss"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync/atomic"
	"time"
)

const (
	cpuMode = 1 << iota
	heapMode
	allocsMode
	mutexMode
	blockMode
	traceMode
	threadCreateMode
	goroutineMode
)

// Profile represents an active profiling session.
type Profile struct {
	// mode holds the type of profiling that will be made
	mode int

	// memProfileRate holds the rate for the memory profile.
	memProfileRate int

	// closer holds a cleanup function that run after each profile
	closer []func()

	// stopped records if a call to profile.Stop has been made
	stopped uint32

	// duration set the profile to stop after the specified duration.
	duration time.Duration
}

// Specs specifies the profile settings by specifies expression.
// like cpu,heap,allocs,mutex,block,trace,threadcreate,goroutine,d:5m,rate:4096.
func Specs(specs string) func(*Profile) {
	return func(p *Profile) {
		for _, spec := range ss.Split(specs, ss.WithIgnoreEmpty(true), ss.WithCase(ss.CaseLower), ss.WithSeps(", ")) {
			switch spec {
			case "cpu":
				p.mode |= cpuMode
			case "heap":
				p.mode |= heapMode
			case "allocs":
				p.mode |= allocsMode
			case "mutex":
				p.mode |= mutexMode
			case "block":
				p.mode |= blockMode
			case "trace":
				p.mode |= traceMode
			case "threadcreate":
				p.mode |= threadCreateMode
			case "goroutine":
				p.mode |= goroutineMode
			default:
				switch {
				case ss.HasPrefix(spec, "d:"):
					if d, err := time.ParseDuration(spec[2:]); err == nil {
						p.duration = d
						continue
					}
				case ss.HasPrefix(spec, "rate:"):
					if r, err := ss.ParseIntE(spec[5:]); err == nil {
						p.memProfileRate = r
						continue
					}
				}
				log.Printf("W! unknown spec: %s", spec)
			}
		}
	}
}

// DefaultMemProfileRate is the default memory profiling rate.
// See also http://golang.org/pkg/runtime/#pkg-variables
const DefaultMemProfileRate = 4096

// Stop stops the profile and flushes any unwritten data.
func (p *Profile) Stop() {
	if !atomic.CompareAndSwapUint32(&p.stopped, 0, 1) {
		// someone has already called close
		return
	}
	for _, closer := range p.closer {
		closer()
	}
	atomic.StoreUint32(&started, 0)
}

// started is non zero if a profile is running.
var started uint32

// Start starts a new profiling session.
// The caller should call the Stop method on the value returned
// to cleanly stop profiling.
func Start(options ...func(*Profile)) (interface{ Stop() }, error) {
	if !atomic.CompareAndSwapUint32(&started, 0, 1) {
		log.Fatal("profile: Start() already called")
	}

	prof := Profile{memProfileRate: DefaultMemProfileRate, duration: 15 * time.Second}
	for _, option := range options {
		option(&prof)
	}

	if prof.mode&cpuMode > 0 {
		f, fcloser, err := createFile("cpu.pprof")
		if err != nil {
			return nil, err
		}
		pprof.StartCPUProfile(f)
		prof.closer = append(prof.closer, func() {
			pprof.StopCPUProfile()
			fcloser()
		})
	}
	if prof.mode&allocsMode > 0 {
		f, fcloser, err := createFile("allocs.pprof")
		if err != nil {
			return nil, err
		}
		old := runtime.MemProfileRate
		runtime.MemProfileRate = prof.memProfileRate
		prof.closer = append(prof.closer, func() {
			pprof.Lookup("allocs").WriteTo(f, 0)
			runtime.MemProfileRate = old
			fcloser()
		})
	}
	if prof.mode&heapMode > 0 {
		f, fcloser, err := createFile("heap.pprof")
		if err != nil {
			return nil, err
		}
		old := runtime.MemProfileRate
		runtime.MemProfileRate = prof.memProfileRate
		prof.closer = append(prof.closer, func() {
			pprof.Lookup("heap").WriteTo(f, 0)
			runtime.MemProfileRate = old
			fcloser()
		})
	}
	if prof.mode&mutexMode > 0 {
		f, fcloser, err := createFile("mutex.pprof")
		if err != nil {
			return nil, err
		}
		runtime.SetMutexProfileFraction(1)
		prof.closer = append(prof.closer, func() {
			if mp := pprof.Lookup("mutex"); mp != nil {
				mp.WriteTo(f, 0)
			}
			runtime.SetMutexProfileFraction(0)
			fcloser()
		})
	}
	if prof.mode&blockMode > 0 {
		f, fcloser, err := createFile("block.pprof")
		if err != nil {
			return nil, err
		}
		runtime.SetBlockProfileRate(1)
		prof.closer = append(prof.closer, func() {
			pprof.Lookup("block").WriteTo(f, 0)
			runtime.SetBlockProfileRate(0)
			fcloser()
		})
	}
	if prof.mode&threadCreateMode > 0 {
		f, fcloser, err := createFile("threadcreation.pprof")
		if err != nil {
			return nil, err
		}
		prof.closer = append(prof.closer, func() {
			if mp := pprof.Lookup("threadcreate"); mp != nil {
				mp.WriteTo(f, 0)
			}
			fcloser()
		})
	}
	if prof.mode&traceMode > 0 {
		f, fcloser, err := createFile("trace.out")
		if err != nil {
			return nil, err
		}
		trace.Start(f)
		prof.closer = append(prof.closer, func() {
			trace.Stop()
			fcloser()
		})
	}
	if prof.mode&goroutineMode > 0 {
		f, fcloser, err := createFile("goroutine.pprof")
		if err != nil {
			return nil, err
		}
		prof.closer = append(prof.closer, func() {
			if mp := pprof.Lookup("goroutine"); mp != nil {
				mp.WriteTo(f, 0)
			}
			fcloser()
		})
	}

	if len(prof.closer) == 0 {
		return nil, fmt.Errorf("no profiles")
	}

	if prof.duration > 0 {
		time.Sleep(prof.duration)
		log.Printf("profile: after %s, stopping profiles", prof.duration)
		prof.Stop()
		return nil, nil
	}

	return &prof, nil
}

func createFile(fn string) (*os.File, func(), error) {
	f, err := os.Create(fn)
	if err != nil {
		log.Printf("E! profile: could not create   %q: %v", fn, err)
		return nil, nil, err
	}
	log.Printf("profile: profiling started, %s", fn)

	fcloser := func() {
		f.Close()
		log.Printf("profile: profiling ended, %s", fn)
	}

	return f, fcloser, err
}
