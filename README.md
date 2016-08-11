# Metrics Capacitor

> Metrics processing engine with ElasticSearch as a backend, like Logstash is for logs :)

## Features

Here you can see the project status at a glance

- [x] (v0.1) switch to Golang :)
- [x] (v0.2) concurrent bulk writer
- [x] (v0.2) standalone writer mode
- [x] (v0.2) TCP listener with influx codec
- [x] (v0.3) TCP listener with graphite codec
- [ ] (v0.4) standalone listener mode
- [ ] (v0.4) logger
- [ ] (v0.4) signal responsiveness
- [ ] (v0.4) safe shutdown (no metric shall be lost)
- [ ] (v1.0) HTTP API
- [ ] (v1.0) HTTP listener with JSON codec
- [ ] (v1.0) aggregator for old metrics
- [ ] (v1.?) StatsD codec

## Prerequisities

- Redis 3.x
- ElasticSearch 2.3
- Go 1.6 (for development)

## Installation

1. Make sure you have Redis and ElasticSearch up and accessible
2. Download the latest [release](https://github.com/metrics-capacitor/metrics-capacitor/releases/latest):
  ```wget -O- -q https://api.github.com/repos/metrics-capacitor/metrics-capacitor/releases/latest | jq -r ".assets[] | select(.name) | .browser_download_url" | xargs sudo wget -O/usr/local/bin/metrics-capacitor```
3. Place the ```metrics-capacitor``` binary into your ```$PATH```


## Configuration

See contents of ```etc/``` directory for details.

## Usage

Just start the Engine by invoking ```metrics-capacitor```

### Options

```
# metrics-capacitor -help
Usage of metrics-capacitor:
  -config string
    	Path to config file (default "/etc/metrics-capacitor/main.conf")
  -cores int
    	Number of cores to use (default all cores)
  -daemonize
    	Run on background
  -version
    	Show version
```

### Metrics
**TODO**


## Development

Everything is handled by Makefile

- ```make prepare``` - create Docker devel environment
- ```make build``` - build Docker image with Metrics Capacitor
- ```make push``` - push built Docker image

You can also manually grab the built binary from ```bin/```
