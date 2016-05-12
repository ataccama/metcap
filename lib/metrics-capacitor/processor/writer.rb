require 'elasticsearch'

module MetricsCapacitor
  module Processor
    class Writer < Core

      def post_init
        @elastic = Elasticsearch::Client.new(
          url: Config.elasticsearch[:urls],
          reload_connections: 100,
          retry_on_failure: Config.elasticsearch[:retry],
          sniffer_timeout: 5,
        )
        logger.debug "Elastic connection set up"

        @redis = Redis.new(url: Config.redis[:url])
        logger.debug "Redis connection set up"

        @exit = false
      end

      def process
        until @exit
          logger.debug "Gathering mertics bulk"
          @metrics = Metrics.new([])
          t = Thread.new do
            while result = @redis.blpop('writer', timeout: Config.writer[:bulk_wait])
              @metrics << Metric.new(result)
            end
          end
          t.join

          if @metrics.empty?
            logger.warn 'No metrics gathered'
          else
            logger.debug "Preparing to write #{@metrics.length} metrics"
            @elastic.bulk(
              index: Config.elasticsearch[:index],
              type: Config.writer[:doc_type],
              body: @metrics.to_elastic,
              fields: ''
            )
          end
        end
      end

      def shutdown
        @exit = true
      end
    end
  end
end
