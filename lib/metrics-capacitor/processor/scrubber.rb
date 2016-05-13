require_relative '../config'
require_relative '../model'

module MetricsCapacitor
  module Processor
    class Scrubber
      include Sidekiq::Worker
      include MetricsCapacitor::Model
      
      sidekiq_options retry: true
      sidekiq_options queue: 'scrubber'

      REDIS = ConnectionPool.new(size: 2) { Redis.new(url: Config.redis[:url]) }

      def process *args
        logger.debug 'Picking redis client from connection-pool'
        REDIS.with do |r|
          logger.debug 'Parsing metrics data'
          metrics = Metrics.new args[0]
          logger.debug 'Sending data to writer'
          r.rpush 'writer', metrics.to_redis
          logger.debug 'Data sent'
        end
      end

    end
  end
end
