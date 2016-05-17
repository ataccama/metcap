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
          metric = nil
          metrics = Metrics.new
          indexing_result = nil
          begin
            while !@exit && metrics.length < Config.writer[:bulk_max] && ( metric = @redis.blpop('writer', timeout: Config.writer[:bulk_wait]) )
              metrics << Metric.new(metric[1])
              metric = nil
            end
          rescue Redis::CannotConnectError, Redis::TimeoutError => e
            logger.error "Cant connect to redis: #{e.message}"
            sleep 1
            retry
          end
          unless metrics.empty?
            logger.info "Writing #{metrics.length} metrics"
            indexing_result = @elastic.bulk(index: Config.elasticsearch[:index], type: Config.writer[:doc_type], body: metrics.to_elastic)
            logger.info "Written #{metrics.length} metrics"
          else
            logger.warn 'No metrics to write :-('
          end
          metrics = nil
          indexing_result = nil
        end
      end

      def shutdown
        @exit = true
      end
    end
  end
end
