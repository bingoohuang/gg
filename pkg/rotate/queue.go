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
	config         *Config
	delayDiscarded *jihe.DelayChan

	wg sync.WaitGroup
}

type Config struct {
	context.Context
	OutChanSize    int    // 通道大小
	AllowDiscarded bool   // 是否允许放弃（来不及写入）
	Append         bool   // 追加模式
	MaxSize        uint64 // 单个文件最大大小
	KeepDays       int    // 保留多少天的日志，过期删除， 0全部, 默认10天
}

type Option func(*Config)

func WithContext(v context.Context) Option { return func(c *Config) { c.Context = v } }
func WithConfig(v *Config) Option          { return func(c *Config) { *c = *v } }
func WithAllowDiscard(v bool) Option       { return func(c *Config) { c.AllowDiscarded = v } }
func WithAppend(v bool) Option             { return func(c *Config) { c.Append = v } }
func WithOutChanSize(v int) Option         { return func(c *Config) { c.OutChanSize = v } }
func WithMaxSize(v uint64) Option          { return func(c *Config) { c.MaxSize = v } }
func WithKeepDays(v int) Option            { return func(c *Config) { c.KeepDays = v } }

// NewQueueWriter creates a new QueueWriter.
// outputPath:
// 1. stdout for the stdout
// 2. somepath/yyyyMMdd.log for the disk file
// 2.1. somepath/yyyyMMdd.log:append for the disk file for append mode
// 2.2. somepath/yyyyMMdd.log:100m for the disk file max 100MB size
// 2.3. somepath/yyyyMMdd.log:100m:append for the disk file max 100MB size and append mode
func NewQueueWriter(outputPath string, options ...Option) *QueueWriter {
	c := createConfig(options)
	s := ParseOutputPath(c, outputPath)
	w := createWriter(s, c)
	p := &QueueWriter{
		queue:  make(chan string, c.OutChanSize),
		writer: w,
		config: c,
	}

	if c.AllowDiscarded {
		p.delayDiscarded = jihe.NewDelayChan(c.Context, func(v interface{}) {
			p.Send(fmt.Sprintf("\n discarded: %d\n", v.(uint32)), false)
		}, 10*time.Second)
	}
	p.wg.Add(1)
	go p.flushing()

	return p
}

func createConfig(options []Option) *Config {
	c := &Config{KeepDays: 10}
	for _, option := range options {
		option(c)
	}

	if c.OutChanSize <= 0 {
		c.OutChanSize = 1000
	}

	if c.Context == nil {
		c.Context = context.Background()
	}

	return c
}

var digits = regexp.MustCompile(`^\d+$`)

func ParseOutputPath(c *Config, outputPath string) string {
	s := ss.RemoveAll(outputPath, ":append")
	if s != outputPath {
		c.Append = true
	}

	if pos := strings.LastIndex(s, ":"); pos > 0 {
		if !digits.MatchString(s[pos+1:]) {
			maxSize, _ := man.ParseBytes(s[pos+1:])
			if maxSize > 0 {
				c.MaxSize = maxSize
			}
			s = s[:pos]
		}
	}
	return s
}

type LfStdout struct{}

func (l LfStdout) Write(p []byte) (n int, err error) {
	return fmt.Fprintf(os.Stdout, "%s\n", bytes.TrimSpace(p))
}

func createWriter(outputPath string, c *Config) io.Writer {
	if outputPath == "stdout" {
		return &LfStdout{}
	}

	w := NewFileWriter(outputPath, c.MaxSize, c.Append)
	if c.KeepDays > 0 {
		go w.daysKeeping(c.KeepDays)
	}

	return w
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

	if !p.config.AllowDiscarded {
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

func (p *QueueWriter) flushing() {
	defer p.wg.Done()
	if c, ok := p.writer.(io.Closer); ok {
		defer c.Close()
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	count := 0
	f, ok := p.writer.(Flusher)

	ctx := p.config.Context
	for {
		select {
		case msg, ok := <-p.queue:
			if !ok {
				return
			}
			_, _ = p.writer.Write([]byte(msg))
			count++
		case <-ticker.C:
			if count > 0 {
				if ok {
					_ = f.Flush()
				}
				count = 0
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *QueueWriter) daysKeeping() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	ctx := p.config.Context

	for {
		select {
		case <-ticker.C:
			p.removeExpiredFiles()
		case <-ctx.Done():
			return
		}
	}
}

func (p *QueueWriter) Close() error {
	if p.config.AllowDiscarded {
		if val := atomic.LoadUint32(&p.discarded); val > 0 {
			p.queue <- fmt.Sprintf("\n#%d discarded", val)
		}
	}
	close(p.queue)
	p.wg.Wait()
	return nil
}

func (p *QueueWriter) removeExpiredFiles() {

}
