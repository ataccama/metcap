require 'excon'
require 'metrics-capacitor/config'
require 'metrics-capacitor/metrics'

MetricsCapacitor::Config.load!
MetricsCapacitor::Config.sidekiq_server_init!
MetricsCapacitor::Config.sidekiq_client_init!

module MetricsCapacitor
  class Worker
    include Sidekiq::Worker
    sidekiq_options retry: true
    sidekiq_options queue: 'metrics'

    case Config.storage_engine
    when :elastic
      require 'elasticsearch'
      CONN = ConnectionPool.new(size: Config.elastic[:connections]) do
        Elasticsearch::Client.new url: Config.elastic_url,
          adapter: :excon,
          reload_connections: 100,
          retry_on_failure: Config.elastic[:retry],
          sniffer_timeout: 5,
          transport_options: {
              persistent: true,
              read_timeout: Config.elastic[:timeout],
              write_timeout: Config.elastic[:timeout],
              connect_timeout: Config.elastic[:timeout],
              tcp_nodelay: true
          }
      end
      include ESProcessor
    when :influx
      CONN = ConnectionPool.new(size: Config.influx[:connections]) do
        Excon.new Config.influx_url,
          persistent: true,
          expects: [204],
          method: :post,
          read_timeout: Config.influx[:timeout],
          write_timeout: Config.influx[:timeout],
          connect_timeout: Config.influx[:timeout],
          idempotent: true,
          retry_limit: Config.influx[:retry],
          tcp_nodelay: true
      end
      include InfluxProcessor
    end

  end

  class InfluxProcessor
    def process *args
      Metrics.new(args[0]).proc_by_slices!(Config.influx[:slices]) do |metrics|
        CONN.with { |influx| influx.request body: metrics.to_influx }
      end
    end
  end

  class ESProcessor
    def process *args
      Metrics.new(args[0]).proc_by_slices!(Config.elastic[:slices]) do |metrics|
        CONN.with do |es|
          es.bulk index: Config.elastic[:index], type: Config.elastic[:type], body: metrics.to_elastic, fields: ''
        end
      end
    end
  end

end
