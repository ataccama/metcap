# Metrics Capacitor

> Metrics processing framework for (but not limited to) Sensu

## Prerequisities

* Ruby >= 2.0.0
* Sensu monitoring framework
* ElasticSearch >= 2.0
* (Optional) Separate Redis for metrics bufferring. *Of course, you can use the same as Sensu, but when it fills-up with data, you monitoring will turn wacko...*

## Install

Just run:

```sudo gem install metrics-capacitor --no-rdoc --no-ri```

...or use your favorite provisioning system

## Config

### Metrics Capacitor
This is the default config:

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
You can start the service from command-line by running ```metrics-capacitor service start```, or you can use this simple config for Upstart:

```
description 'Metrics Capacitor'
start on virtual-filesystems
stop on runlevel [06]
respawn
limit nofile 65550 65550
console log
exec metrics-capacitor run
```

### Writing Sensu plugins

Same rules apply as for Graphite Sensu plugins. There are a few differrences:

1. The Class to inherit from is ```Sensu::Plugin::Metric::CLI::MetricsCapacitor```
2. The ```-s```/```--scheme``` parameter does not apply any more. Use ```-T <field>=<value>``` to tag you data points. You can specify this parameter multiple times.
3. The resulting data does not flow into Sensu itself. Sensu only gets result regarding (un)successful execution of the plugin

### Reusing existing Graphite plugins

*TODO*

### Limitations

*TODO*
