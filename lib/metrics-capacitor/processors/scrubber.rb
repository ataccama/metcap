module Processors
  module Scrubber

    CONN = ConnectionPool.new(size: Config.elastic[:connections]) do

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
