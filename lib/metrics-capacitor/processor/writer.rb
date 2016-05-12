module MetricsCapacitor
  module Processor
    module Writer

      ES = ConnectionPool.new(size: MetricsCapacitor::Config.elasticsearch[:connections]) do
        Elasticsearch::Client.new(url: MetricsCapacitor::Config.elasticsearch[:url], adapter: :excon, reload_connections: 100, retry_on_failure: MetricsCapacitor::Config.elasticsearch[:retry], sniffer_timeout: 5, transport_options: { persistent: true, read_timeout: MetricsCapacitor::Config.elasticsearch[:timeout], write_timeout: MetricsCapacitor::Config.elasticsearch[:timeout], connect_timeout: MetricsCapacitor::Config.elasticsearch[:timeout], tcp_nodelay: true })
      end

      def process *args
        Metrics.new(args[0]).proc_by_slices!(MetricsCapacitor::Config.elasticsearch[:bulk_max]) do |metrics|
          ES.with do |es|
            es.bulk index: MetricsCapacitor::Config.elastic[:index],
              type: MetricsCapacitor::Config.elastic[:type],
              body: metrics.to_elastic,
              fields: ''
          end
        end
      end

    end
  end
end
