# MetCap

> Metrics processing engine with ElasticSearch as a backend, like Logstash is for logs :)

**Check the [project wiki](https://github.com/blufor/metcap/wiki) for docs!**

**Main features:**
- targeting **1mio+ metrics per sec**
- full **multicore** support
- **listeners** with configurable *codecs*
  - Graphite
  - InfluxDB ([#22](https://github.com/blufor/metcap/issues/22))
  - OpenTSDB ([#24](https://github.com/blufor/metcap/issues/24))
- easy listener **load-balancing** (ie. via HAProxy)
- **transport** implements configurable backends for **multi-host scaling**
  - Go Channel
  - Redis
  - AMQP
  - NATS ([#23](https://github.com/blufor/metcap/issues/23))
- ElasticSearch bulk **writer**
  - simple **data layer scalability** (via ElasticSearch clustering)
- console/syslog **logger**
- use [Grafana](http://grafana.org) as a front-end or write your own ElasticSearch queries :wink:

----------------------------------------------------------------------

Development has been supported by: [Kiwi.com](http://www.kiwi.com/), [Etnetera Group](http://www.etneteragroup.com/), [NeuronAD](http://www.neuronad.com/), blufor's family
