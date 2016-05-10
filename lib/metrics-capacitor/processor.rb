require 'excon'
require 'sidekiq'
require 'sidekiq/logging'
require 'syslog'
require 'log4r'
require 'log4r/configurator'
require 'log4r/outputter/syslogoutputter'
require_relative 'config'
require_relative 'metrics'

module MetricsCapacitor

  Config.load!
  Sidekiq.configure_server do |config|
    config.redis = { url: Config.redis_url }
    Sidekiq::Logging.logger = Log4r::Logger.new 'sidekiq'
    Sidekiq::Logging.logger.outputters = Config.syslog ? Log4r::SyslogOutputter.new('sidekiq', ident: 'metrics-capacitor') : Log4r::Outputter.stdout
    Sidekiq::Logging.logger.level = Log4r::INFO
  end
  Sidekiq.configure_client do |config|
    config.redis = { url: Config.redis_url }
  end

  module Processors
    module Influx
      CONN = ConnectionPool.new(size: Config.influx[:connections]) do
        Excon.new(Config.influx_url, persistent: true, expects: [204], method: :post, read_timeout: Config.influx[:timeout], write_timeout: Config.influx[:timeout], connect_timeout: Config.influx[:timeout], idempotent: true, retry_limit: Config.influx[:retry], tcp_nodelay: true)
      end

      def process *args
        Metrics.new(args[0]).proc_by_slices!(Config.influx[:slices]) do |metrics|
          CONN.with { |influx| influx.request body: metrics.to_influx }
        end
      end
    end

    module Elastic
      require 'elasticsearch'
      CONN = ConnectionPool.new(size: Config.elastic[:connections]) do
        Elasticsearch::Client.new(url: Config.elastic_url, adapter: :excon, reload_connections: 100, retry_on_failure: Config.elastic[:retry], sniffer_timeout: 5, transport_options: { persistent: true, read_timeout: Config.elastic[:timeout], write_timeout: Config.elastic[:timeout], connect_timeout: Config.elastic[:timeout], tcp_nodelay: true })
      end

      def process *args
        Metrics.new(args[0]).proc_by_slices!(Config.elastic[:slices]) do |metrics|
          CONN.with do |es|
            es.bulk index: Config.elastic[:index],
              type: Config.elastic[:type],
              body: metrics.to_elastic,
              fields: ''
          end
        end
      end
    end
  end

  class Processor
    include Sidekiq::Worker
    include Kernel.const_get "MetricsCapacitor::Processors::#{Config.storage_engine.to_s.capitalize}"
    sidekiq_options retry: true
    sidekiq_options queue: 'metrics'
  end

end
