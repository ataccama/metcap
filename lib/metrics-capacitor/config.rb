require 'yaml'
class ::Hash
  def deep_merge(second)
    merger = proc { |key, v1, v2| Hash === v1 && Hash === v2 ? v1.merge(v2, &merger) : v2 }
    self.merge(second, &merger)
  end
end

module MetricsCapacitor
  module Config

    extend self
    attr_reader :_cfg

    def load!
      @_cfg = {
        syslog: false,
        debug: false,
        redis: {
          url: 'redis://127.0.0.1:6379/0',
          timeout: 5,
        },
        elasticsearch: {
          urls: ['http://localhost:9200/'],
          index: 'metrics',
          timeout: 10,
          connections: 2,
        },
        scrubber: {
          threads: 4,
          processes: 2,
          retry: 3,
          tags: {}
        },
        writer: {
          processes: 2,
          doc_type: 'actual',
          bulk_max: 5000,
          bulk_wait: 10,
          ttl: '1w'
        },
        aggregator: {
          doc_type: 'aggregated',
          aggregate_by: 600, # seconds
          optimize_indices: true,
        }
      }
      @_cfg = @_cfg.deep_merge YAML.load_file('/etc/metrics-capacitor.yaml') if File.exists? '/etc/metrics-capacitor.yaml'
    end

    def method_missing (name, *args, &block)
      return @_cfg[name.to_sym] if @_cfg[name.to_sym] != nil
      fail(NoMethodError, "Unknown configuration section Config.#{name}", caller)
    rescue NoMethodError => e
      $stderr.puts "ERROR config: #{e.class}: #{e.message}"
      $stderr.puts e.backtrace.map { |l| "ERROR config: #{l}\n" }.join
      exit! 1
    end

  end
end
