package metcap

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Syslog      bool
	Debug       bool
	ReportEvery configDuration `toml:"report_every"`
	Transport   TransportConfig
	Listener    map[string]ListenerConfig
	Writer      WriterConfig
	Aggregator  AggregatorConfig
}

type TransportConfig struct {
	Type             string
	BufferSize       int    `toml:"buffer_size"`
	RedisURL         string `toml:"redis_url"`
	RedisTimeout     int    `toml:"redis_timeout"`
	RedisWait        int    `toml:"redis_wait"`
	RedisRetries     int    `toml:"redis_retries"`
	RedisConnections int    `toml:"redis_connections"`
	RedisQueue       string `toml:"redis_queue"`
	AMQPURL          string `toml:"amqp_url"`
	AMQPTag          string `toml:"amqp_tag"`
	AMQPTimeout      int    `toml:"amqp_timeout"`
	AMQPWorkers      int    `toml:"amqp_workers"`
}

type ListenerConfig struct {
	Port        int
	Protocol    string
	Codec       string
	Decoders    int
	MutatorFile string `toml:"mutator_file"`
}

type WriterConfig struct {
	URLs        []string       `toml:"urls"`
	Timeout     int            `toml:"timeout"`
	Concurrency int            `toml:"concurrency"`
	BulkMax     int            `toml:"bulk_max"`
	BulkWait    configDuration `toml:"bulk_wait"`
	Index       string         `toml:"index"`
	DocType     string         `toml:"doc_type"`
}

type AggregatorConfig struct{}

type configDuration struct {
	time.Duration
}

func (d *configDuration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// ReadConfig
//
func ReadConfig(configfile *string) Config {
	if _, err := os.Stat(*configfile); err != nil {
		fmt.Println("Can't read configfile")
		os.Exit(1)
	}

	var config Config
	if _, err := toml.DecodeFile(*configfile, &config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return config
}
