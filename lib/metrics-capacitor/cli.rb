require 'thor'

module MetricsCapacitor
  class CLI < Thor
    package_name 'Metrics Capacitor'

    desc 'engine', 'Start the engine :-)'
    long_desc <<-LONGDESC
      Usage: metrics-capacitor engine

      all options are defined in /etc/metrics-capacitor.yaml
    LONGDESC
    def engine
      require 'metrics-capacitor/engine'
      Engine.new.run!
    end

    desc 'graphite', 'Send Graphite data'
    option :tag_map, type: :string, required: true, aliases: '-m'
    option :add_tag, type: :hash, aliases: '-t'
    option :debug, type: :boolean, aliases: '-d'
    option :counter_match, type: :string
    def graphite
      require 'metrics-capacitor/utils/graphite'
      Utils::Graphite.new(options).run!
    end

    # desc 'status', 'Report the state (TODO)'
    # def status
    #   # TODO ...
    # end
    #
    # desc 'aggregate', 'Manually trigger aggregator run (TODO)'
    # def aggregate
    #   # TODO ...
    # end
    #
    # desc 'expunge', 'Expunge old data (TODO)'
    # def expunge
    #   # TODO ...
    # end
    #
    # desc 'optimize', 'Optimize ElasticSearch indices (TODO)'
    # def optimize
    #   # TODO ...
    # end

  end
end
