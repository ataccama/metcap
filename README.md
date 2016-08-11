# Metrics Capacitor

> Metrics processing engine with ElasticSearch backend, like Logstash is for logs :)

**Work in progress...**

## Features
- [x] Concurrent bulk writer
- [x] TCP listener with influx codec
- [x] TCP listener with graphite codec
- [ ] HTTP listener with JSON codec
- [ ] UDP listener with StatsD codec
- [ ] HTTP API
- [ ] logger
- [ ] Signal responsiveness
- [ ] safe shutdown (no metric shall be lost)

## Prerequisities

- Redis 3.x
- ElasticSearch 2.3
- Go 1.6 (for development)

## Usage

**TODO**

## Configuration

See contents of ```etc/``` directory.

## Development

Everything is handled by Makefile

- ```make prepare``` - create Docker devel environment
- ```make build``` - build Docker image with Metrics Capacitor
- ```make push``` - push built Docker image

You can also manually grab the built binary from ```bin/```
