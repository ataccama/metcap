require 'thor'
require_relative 'process'

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

    desc 'status', 'Report the state'
    def status
      Config.sidekiq_init_client!
      # TODO ...
    end
  end
end
