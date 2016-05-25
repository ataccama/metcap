require 'thor'

module MetricsCapacitor
  class CLI < Thor
    package_name 'Metrics Capacitor'

    desc 'engine', 'Start the engine :-)'
    def engine
      require 'metrics-capacitor/engine'
      Engine.new.run!
    end

    desc 'graphite', 'Send Graphite data'
    long_desc <<-LONGDESC
      `metrics-capacitor graphite`
    LONGDESC
    option :tag_map, required: true
    option :add_tags, type: :hash
    option :debug, type: :boolean
    option :counter_match
    def graphite
      require 'metrics-capacitor/utils/graphite'
      Utils::Graphite.new(options).run!
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
