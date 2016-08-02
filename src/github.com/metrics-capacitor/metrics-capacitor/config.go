package metcap

import (
  "os"
  "fmt"

  "github.com/BurntSushi/toml"
)

type Config struct {
  Syslog        bool
  Debug         bool
  Redis         RedisConfig
  Listener      map[string]ListenerConfig
  Writer        WriterConfig
  Aggregator    AggregatorConfig
}

type RedisConfig struct {
  Socket      string
  Address     string
  DB          int
  Timeout     int
  Connections int
  Queue       string
}

type ListenerConfig struct {
  Protocol  string
  Port      int
}

type WriterConfig struct {
  Urls        []string
  Timeout     int
  Concurrency int
  BulkMax     int
  BulkWait    int
  Index       string
  DocType     string
  Ttl         int
}

type AggregatorConfig struct {}

// Read config file
//
func ReadConfig(configfile *string) Config {
	if _, err := os.Stat(*configfile); err != nil {
		fmt.Println("Can't read %s", *configfile)
    os.Exit(1)
	}

	var config Config
	if _, err := toml.DecodeFile(*configfile, &config); err != nil {
		fmt.Println(err)
    os.Exit(1)
	}
	return config
}
