module Processors
  module Writer

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
