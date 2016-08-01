package metcap

import (
  "github.com/BurntSushi/toml"
  "os"
  "fmt"
)

// Main configuration struct
//
type Config struct {
  Syslog  bool
  Debug   bool
  Redis   struct {
    Url       string
    Timeout   int
    MaxIdle   int
    MaxActive int
  }
  Elasticsearch struct {
    Urls        []string
    Timeout     int
    Connections int
    Index       string
  }
  Listener []ListenerConfig
  Scrubber struct {
    Threads int
  }
  Writer struct {
    Threads   int
    DocType   string
    BulkMax   int
    BulkWait  int
    Ttl       int
  }
  Aggregator struct {
    DocType         string
    AggregateBy     int
    OptimizeIndices bool
  }
}

// Listener configuration sub-type
//
type ListenerConfig struct {
  Type  string
  Port  int
}

// Read config file
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
