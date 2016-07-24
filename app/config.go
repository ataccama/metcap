// Main configuration struct
//
type Config struct {
  Syslog  bool
  Debug   bool
  Redis   struct {
    Url     string
    Timeout int
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

// Read config file
func ReadConfig() Config {
	var configfile = flags.Configfile
	if _, err := os.Stat(configfile); err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}
	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
	log.Print(config.Index)
	return config
}

// Listener configuration sub-type
//
type ListenerConfig struct {
  Type  string
  Port  int
}
