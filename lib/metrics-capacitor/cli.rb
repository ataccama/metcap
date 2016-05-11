require 'thor'
require_relative 'engine'

module MetricsCapacitor
  class CLI < Thor
    package_name 'Metrics Capacitor'

    desc 'engine', 'Start the engine :-)'
    def engine
      p = Engine.instance
      p.run_scrubber!
      p.run_writer!
      p.run_aggregator!
      p.wait
    end

    desc 'status', 'Report the state (TODO)'
    def status
      # TODO ...
    end

    desc 'aggregate', 'Manually trigger aggregator run (TODO)'
    def aggregate
      # TODO ...
    end

    desc 'expunge', 'Expunge old data (TODO)'
    def expunge
      # TODO ...
    end

    desc 'optimize', 'Optimize ElasticSearch indices (TODO)'
    def optimize
      # TODO ...
    end

  end
end
