# MetCap

> Metrics processing engine with ElasticSearch as a backend, like Logstash is for logs :)

Features:
- Written in Golang... for speed ;)
- Modular design
  - transport (Go Channel/Redis/AMQP/NATS)
  - listener codecs (Graphite, InfluxDB)
  - ElasticSearch writer
- Connection pooling
- Scalability
  - full multicore support
  - easy load-balancing
  - simple data layer scalability

**Check the [project wiki](https://github.com/blufor/metcap/wiki) for docs!**


----------------------------------------------------------------------

Development has been supported by: [Kiwi.com](http://www.kiwi.com/), [Etnetera Group](http://www.etneteragroup.com/), [NeuronAD](http://www.neuronad.com/), blufor's family
