module MetricsCapacitor
  module Processor
    class Scrubber
      include Sidekiq::Worker

      REDIS = ConnectionPool.new(size: 2) { Redis::Client.new() }

      def process *args
        REDIS.with { |r| r.rpush 'writer', Metrics.new(args[0]).to_redis }
      end

    end
  end
end
