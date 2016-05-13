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
        logger.debug 'Elastic connection set up'

        @redis = Redis.new(url: Config.redis[:url])
        logger.debug 'Redis connection set up'

        @exit = false
      end

      def process
        logger.debug 'Randomizing startup time'
        sleep rand(Config.writer[:bulk_wait])
        until @exit
          logger.debug 'Gathering mertics bulk'
          metrics = Metrics.new
          while !@exit && metrics.length < Config.writer[:bulk_max] && ( result = @redis.blpop('writer', timeout: Config.writer[:bulk_wait]) )
            metrics << Metric.new(result[1])
          end

          if metrics.empty?
            logger.warn 'No metrics to write'
          else
            logger.info "Preparing to write #{metrics.length} metrics"
            @elastic.bulk(index: Config.elasticsearch[:index], type: Config.writer[:doc_type], body: metrics.to_elastic)
          end
          metrics = nil
        end
      end

      def shutdown
        @exit = true
      end
    end
  end
end
