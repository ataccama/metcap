# Metrics Capacitor

> Metrics processing engine with ElasticSearch as a backend, like Logstash is for logs :)

## Features

- [x] switch to Golang :) (v0.1)
- [x] concurrent bulk writer (v0.2)
- [x] standalone writer mode (v0.2)
- [x] TCP listener with influx codec (v0.2)
- [ ] TCP listener with graphite codec (v0.3)
- [ ] logger (v0.3)
- [ ] signal responsiveness (v0.3)
- [ ] safe shutdown (no metric shall be lost) (v0.3)
- [ ] HTTP API (v1.0)
- [ ] HTTP listener with JSON codec (v1.0)
- [ ] aggregator for old metrics (v1.0)
- [ ] StatsD codec (v1.?)

## Prerequisities

- Redis 3.x
- ElasticSearch 2.3
- Go 1.6 (for development)

## Installation

**TODO**

## Configuration

See contents of ```etc/``` directory.

## Usage

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
