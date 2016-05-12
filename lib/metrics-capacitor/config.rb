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
        debug: true,
        redis: {
          url: 'redis://127.0.0.1:6379/0'
        },
        elasticsearch: {
          urls: ['http://localhost:9200/'],
          index: 'metrics',
          timeout: 10,
        },
        scrubber: {
          threads: 16,
          processes: 1,
          retry: 3,
          tags: {},
          worker_path: worker_path
        },
        writer: {
          processes: 2,
          doc_type: 'actual',
          bulk_max: 1000,
          bulk_wait: 15,
          retry: true
        },
        aggregator: {
          doc_type: 'aggregated',
          aggregate_by: 600, # seconds
          optimize_indexes: true,
          expunge_after: 2678400 # 31 days
        }
      }
      @_cfg = @_cfg.deep_merge YAML.load_file('/etc/metrics-capacitor.yaml') if File.exists? '/etc/metrics-capacitor.yaml'
    end

    def worker_path
      File.expand_path('..', __FILE__) + '/processor/scrubber.rb'
    end

    def method_missing (name, *args, &block)
      return @_cfg[name.to_sym] if @_cfg[name.to_sym] != nil
      fail(NoMethodError, "Unknown configuration section Config.#{name}", caller)
    rescue NoMethodError => e
      $stderr.puts " ERROR config: #{e.class}: #{e.message}"
      $stderr.puts " ERROR config: #{e.backtrace.first}"
      exit! 1
    end

  end
end
