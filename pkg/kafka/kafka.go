package kafka

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/Shopify/sarama"
)

type ProducerConfig struct {
	Topic        string
	Version      string
	Brokers      []string
	Codec        string
	Sync         bool
	RequiredAcks sarama.RequiredAcks

	TlsConfig TlsConfig
	Context   context.Context
}

type PublishResponse struct {
	Result interface{}
	Topic  string
}

type Options struct {
	MessageKey string
	Headers    map[string]string
}

func (o *Options) Fulfil(msg *sarama.ProducerMessage) {
	if len(o.Headers) > 0 {
		msg.Headers = make([]sarama.RecordHeader, 0, len(o.Headers))
		for k, v := range o.Headers {
			msg.Headers = append(msg.Headers, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
		}
	}

	if len(o.MessageKey) > 0 {
		msg.Key = sarama.StringEncoder(o.MessageKey)
	}
}

type OptionFn func(*Options)

func WithKey(key string) OptionFn     { return func(options *Options) { options.MessageKey = key } }
func WithHeader(k, v string) OptionFn { return func(options *Options) { options.Headers[k] = v } }

type asyncProducer struct {
	sarama.AsyncProducer
}

func (p asyncProducer) SendMessage(msg *sarama.ProducerMessage) (interface{}, error) {
	p.Input() <- msg
	return AsyncProducerResult{Enqueued: true}, nil
}

type AsyncProducerResult struct {
	Enqueued    bool
	ContextDone bool
}

type SyncProducerResult struct {
	Partition int32
	Offset    int64
}

type syncProducer struct {
	sarama.SyncProducer
}

func (p syncProducer) SendMessage(msg *sarama.ProducerMessage) (interface{}, error) {
	partition, offset, err := p.SyncProducer.SendMessage(msg)
	return SyncProducerResult{Partition: partition, Offset: offset}, err
}

type producer interface {
	SendMessage(*sarama.ProducerMessage) (interface{}, error)
}

type Producer struct {
	producer
	io.Closer
}

func (p Producer) Publish(topic string, vv []byte, optionsFns ...OptionFn) (*PublishResponse, error) {
	options := &Options{Headers: make(map[string]string)}
	for _, f := range optionsFns {
		f(options)
	}

	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(vv)}
	options.Fulfil(msg)

	result, err := p.producer.SendMessage(msg)
	if err != nil {
		return nil, err
	}

	return &PublishResponse{Result: result, Topic: msg.Topic}, nil
}

func (c ProducerConfig) NewProducer() (*Producer, error) {
	// For the data collector, we are looking for strong consistency semantics.
	// Because we don't change the flush settings, sarama will try to produce messages
	// as fast as possible to keep latency low.
	sc := sarama.NewConfig()
	if err := ParseVersion(sc, c.Version); err != nil {
		return nil, err
	}

	sc.Producer.Compression = ParseCodec(c.Codec)
	sc.Producer.RequiredAcks = c.RequiredAcks
	sc.Producer.Retry.Max = 3 // Retry up to x times to produce the message
	sc.Producer.MaxMessageBytes = int(sarama.MaxRequestSize)
	sc.Producer.Return.Successes = true
	if tc := c.TlsConfig.Create(); tc != nil {
		sc.Net.TLS.Config = tc
		sc.Net.TLS.Enable = true
	}

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.
	if c.Sync {
		p, err := sarama.NewSyncProducer(c.Brokers, sc)
		if err != nil {
			return nil, fmt.Errorf("failed to start Sarama SyncProducer, %w", err)
		}
		return &Producer{producer: &syncProducer{SyncProducer: p}}, nil
	}

	sc.Producer.Return.Errors = true
	p, err := sarama.NewAsyncProducer(c.Brokers, sc)
	if err != nil {
		return nil, fmt.Errorf("failed to start Sarama NewAsyncProducer, %w", err)
	}
	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for {
			select {
			case <-c.Context.Done():
				return
			case <-p.Successes():
			case err := <-p.Errors():
				log.Println("Failed to write access log entry:", err)
			}
		}
	}()

	return &Producer{producer: &asyncProducer{AsyncProducer: p}}, nil
}
