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
    long_desc <<-LONGDESC
    This triggers Graphite metrics transformer. The data is awaited on STDIN.
    AVOID EMPTY LINES AT THE END OF OUTPUT AT ALL COSTS

    TAG_MAP option is mandatory and controls the behavior of graphite path
    transformation into Elasticsearch fields, we call @tags and @name. Tag map must
    correspond to the graphite data path.

    There are 3 acceptable values: Integer (any sane positive whole number),
    String (anything that Ruby parser can't convert to Integer) and special string `_`,
    Integers are used as an array field indexes that are joined by `:` into @name field.
    Strings are used as as @tags subfield names. And finally `_` is used if you want to
    ignore the value on the position.

    EXAMPLE:

    Graphite path `server.node10.redis.hz` and TAG_MAP (-m) `_.host.0.1`
    result metric with @name: `redis:hz` and @tags: `{ host: node10 }`

    LONGDESC
    option :tag_map, type: :string, required: true, aliases: '-m'
    option :add_tag, type: :hash, aliases: '-t'
    option :debug, type: :boolean, aliases: '-d'
    option :counter_match, type: :string
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
