# Metrics Capacitor

> Metrics processing engine with ElasticSearch as a backend, like Logstash is for logs :)

Development has been supported by:
- [EtNetera Group](http://www.etneteragroup.com/)
- [NeuronAD LTD](http://www.neuronad.com/)

----------------------------------------------------------------------

## Table of contents

* [Features](#features)
* [Prerequisities](#prerequisities)
* [Installation](#installation)
* [Configuration](#configuration)
* [Usage](#usage)
  * [Options](#options)
  * [Metrics](#metrics)
    * [Influx data format](#influx-data-format)
    * [Graphite data format](#graphite-data-format)
      * [Example](#example)
* [Development](#development)

## Features

- Written in Golang... for speed ;)
- Graphite + InfluxDB line listeners (TCP)
- Modular design
  - Transport (Go Channel/Redis/AMQP)
  - Listeners (Graphite, InfluxDB)
  - Writer
- Connection pooling
- Scalability
  - Multiple-core use
  - Multiple hosts can be HAProxied ;)
  - Data layer scalability is provided natively by ElasticSearch

## Prerequisities

- ElasticSearch 2.3
- Redis 3.x or RabbitMQ with AMQP 0.9.1 (optional, for scalability)
- Go 1.6 (for development)

## Installation

1. Make sure you have Redis and ElasticSearch up and accessible
2. Download the latest [release](https://github.com/metrics-capacitor/metrics-capacitor/releases/latest) into ```/usr/local/bin```:
  ```wget -O- -q https://api.github.com/repos/metrics-capacitor/metrics-capacitor/releases/latest | jq -r ".assets[] | select(.name) | .browser_download_url" | xargs sudo wget -O/usr/local/bin/metrics-capacitor && sudo chmod a+x /usr/local/bin/metrics-capacitor```


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

#### Influx data format

- metric without timestamp: ```metric_name key1=foo,key2=bar 10```
- metric with timestamp: ```metric_name key1=foo,key2=bar 10 1470929084```

#### Graphite data format

- metric without timestamp: ```some.path.to.metric 10```
- metric with timestamp: ```some.path.to.metric 10 1470929084```

All paths are matched against rules defined in ```mutator_file```. Each line in the file represents one rule. The line has two values separated by ```|||```, the first is RegEx matching the Graphite path, the second describes mapping of values to field names and metric name. The mapping pattern can have following values:
- ```-```: name of the leaf in path is ommitted
- ```(int)```: name of the leaf path will be part of ```name``` field. If you use multiple numbers then the ```name``` field will contain all those leaf names separated by ```:```
- ```(string)``` : name of the leaf will become value for the key specified by the ```(string)```

Additionally you can use a special flag ```+``` at the end of each line. It can be used in two ways:
- **standalone** - ie. ```test.+``` - this will capture all the remaining path leaves and put them into *name* while putting the top leaf into *test* field.
- **with string** - ie. ```1.data+``` - this will capture all the remaining path leaves and put them into *data* field while putting the top leaf into *name*.

##### Example

Metric data: ```stats.counter.test.alpha.rate 10```


- Mutator rule: ```^stats\.|||-.type.section.+```
- Resulting metric: ```{"name": "alpha:rate", "section": "test", "value": 10, "type": "counter", "@timestamp": "..."}```


- Mutator rule: ```^stats\.|||-.type.1.statistic+```
- Resulting metric: ```{"name": "test", "statistic": "alpha:rate", "value": 10, "type": "counter", "@timestamp": "..."}```


## Development

Everything is handled by Makefile

- ```make prepare``` - create Docker devel environment
- ```make build``` - build Docker image with Metrics Capacitor
- ```make push``` - push built Docker image

You can also manually grab the built binary from ```bin/```
