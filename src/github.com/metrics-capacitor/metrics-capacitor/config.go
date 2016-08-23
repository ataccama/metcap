package metcap

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Syslog     bool
	Debug      bool
	Buffer     BufferConfig
	Listener   map[string]ListenerConfig
	Writer     WriterConfig
	Aggregator AggregatorConfig
}

type BufferConfig struct {
	Socket      string
	Address     string
	DB          int
	Timeout     int
	Wait        int
	Connections int
	Queue       string
}

type ListenerConfig struct {
	Port        int
	Protocol    string
	Codec       string
	MutatorFile string	`toml:"mutator_file"`
}

type WriterConfig struct {
	Urls        []string
	Timeout     int
	Concurrency int
	BulkMax     int 		`toml:"bulk_max"`
	BulkWait    int			`toml:"bulk_wait"`
	Index       string
	DocType     string	`toml:"doc_type"`
	Ttl         int
}

type AggregatorConfig struct{}

// Read config file
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
