# MetCap

> Metrics processing engine with ElasticSearch as a backend, like Logstash is for logs :)

> *Formerly metrics-capacitor*

[Check the project wiki for docs](https://github.com/blufor/metcap/wiki)

# Features

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

----------------------------------------------------------------------

Development has been supported by: [Kiwi.com](http://www.kiwi.com/), [Etnetera Group](http://www.etneteragroup.com/), [NeuronAD](http://www.neuronad.com/)
