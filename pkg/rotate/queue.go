package rotate

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bingoohuang/gg/pkg/jihe"
	"github.com/bingoohuang/gg/pkg/man"
	"github.com/bingoohuang/gg/pkg/ss"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// QueueWriter output parsed http messages
type QueueWriter struct {
	queue  chan string
	writer io.Writer

	discarded      uint32
	allowDiscard   bool
	delayDiscarded *jihe.DelayChan

	wg sync.WaitGroup
}

// NewQueueWriter creates a new QueueWriter.
// outputPath: 1. stdout for the stdout
//             2. somepath/yyyyMMdd.log for the disk file
//             2.1. somepath/yyyyMMdd.log:append for the disk file for append mode
//             2.2. somepath/yyyyMMdd.log:100m for the disk file max 100MB size
//             2.3. somepath/yyyyMMdd.log:100m:append for the disk file max 100MB size and append mode
func NewQueueWriter(ctx context.Context, outputPath string, outChanSize uint, allowDiscarded bool) *QueueWriter {
	s, appendMode, maxSize := ParseOutputPath(outputPath)
	w := createWriter(s, maxSize, appendMode)
	p := &QueueWriter{
		queue:        make(chan string, outChanSize),
		writer:       w,
		allowDiscard: allowDiscarded,
	}

	if allowDiscarded {
		p.delayDiscarded = jihe.NewDelayChan(ctx, func(v interface{}) {
			p.Send(fmt.Sprintf("\n discarded: %d\n", v.(uint32)), false)
		}, 10*time.Second)
	}
	p.wg.Add(1)
	go p.printBackground(ctx)
	return p
}

var digits = regexp.MustCompile(`^\d+$`)

func ParseOutputPath(outputPath string) (string, bool, uint64) {
	s := ss.RemoveAll(outputPath, ":append")
	appendMode := s != outputPath
	maxSize := uint64(0)
	if pos := strings.LastIndex(s, ":"); pos > 0 {
		if !digits.MatchString(s[pos+1:]) {
			maxSize, _ = man.ParseBytes(s[pos+1:])
			s = s[:pos]
		}
	}
	return s, appendMode, maxSize
}

type LfStdout struct{}

func (l LfStdout) Write(p []byte) (n int, err error) {
	return fmt.Fprintf(os.Stdout, "%s\n", bytes.TrimSpace(p))
}

func createWriter(outputPath string, maxSize uint64, append bool) io.Writer {
	if outputPath == "stdout" {
		return &LfStdout{}
	}

	return NewFileWriter(outputPath, maxSize, append)
}

func (p *QueueWriter) Send(msg string, countDiscards bool) {
	if msg == "" {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("W! Recovered %v", err)
		}
	}() // avoid write to closed p.queue

	if !p.allowDiscard {
		p.queue <- msg
		return
	}

	select {
	case p.queue <- msg:
	default:
		if countDiscards {
			p.delayDiscarded.Put(atomic.AddUint32(&p.discarded, 1))
		}
	}
}

func (p *QueueWriter) printBackground(ctx context.Context) {
	defer p.wg.Done()
	if c, ok := p.writer.(io.Closer); ok {
		defer c.Close()
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-p.queue:
			if !ok {
				return
			}
			_, _ = p.writer.Write([]byte(msg))
		case <-ticker.C:
			if f, ok := p.writer.(Flusher); ok {
				_ = f.Flush()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *QueueWriter) Close() error {
	if p.allowDiscard {
		if val := atomic.LoadUint32(&p.discarded); val > 0 {
			p.queue <- fmt.Sprintf("\n#%d discarded", val)
		}
	}
	close(p.queue)
	p.wg.Wait()
	return nil
}
