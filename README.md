# Metrics Capacitor

> Metrics processing framework for (but not limited to) Sensu

*This project is currently under active development. That means it doesn't work properly... **YET** :)*

## Prerequisities

* Ruby >= 2.0.0

## Install

Please refer to the README of the appropriate module:
* [engine](https://github.com/metrics-capacitor-engine)
* [utils](https://github.com/metrics-capacitor-utils)

## Config

Configuration file applies to all Metrics Capacitor modules. This is the default config hash in YAML format:

```yaml
:debug: false
:syslog: false
:concurrency: 16
:storage_engine: :elastic # either :elastic or :influx
# :sidekiq_path: # YOU BETTER KNOW WHAT YOU DO! ;)
:redis:
  :host: 127.0.0.1
  :port: 6379
  :db: 0
:influx: # required for :influx storage_engine
  :ssl: false
  :host: 127.0.0.1
  :port: 8086
  :path: ''
  :db: metrics
  :timeout: 10
  :slice: 1000
  :retry: 3
  :connections: 4
:elastic: # required for :elastic storage_engine
  :ssl: false
  :host: 127.0.0.1
  :port: 9200
  :path: ''
  :index: metrics
  :type: fresh
  :timeout: 10
  :slice: 5000
  :retry: 3
  :connections: 4
```

Any value can be overloaded in ```/etc/metrics-capacitor.yaml```.

### Sensu Clients

*TODO*

## Use

### Service
You can start the service from command-line by running ```metrics-capacitor engine```, or you can use this simple config for Upstart:

```
description 'Metrics Capacitor'
start on virtual-filesystems
stop on runlevel [06]
respawn
limit nofile 65550 65550
console log
exec metrics-capacitor engine
```

### Writing Sensu metric plugins

Same rules apply as for Graphite Sensu plugins. There are a few differrences:

1. The Class to inherit from is ```Sensu::Plugin::Metric::CLI::MetricsCapacitor```
2. The ```-s```/```--scheme``` parameter does not apply any more. Use ```-T <field>=<value>``` to tag you data points. You can specify this parameter multiple times.
3. The resulting data does not flow into Sensu itself. Sensu only gets result regarding (un)successful execution of the plugin

### Reusing existing Graphite plugins

*TODO*



## Limitations

*TODO*
