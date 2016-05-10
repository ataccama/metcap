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
        concurrency: 16,
        storage_engine: :elastic,
        sidekiq_path: `/bin/which sidekiq`.chomp.to_s,
        redis: {
          host: '127.0.0.1',
          port: 6379,
          db: 0
        },
        influx: {
          ssl: false,
          host: '127.0.0.1',
          port: 8086,
          path: '',
          db: 'metrics',
          timeout: 10,
          slice: 1000,
          retry: 3,
          connections: 4
        },
        elastic: {
          ssl: false,
          host: '127.0.0.1',
          port: 9200,
          path: '',
          index: 'metrics',
          type: 'fresh',
          timeout: 10,
          slice: 5000,
          retry: 3,
          connections: 4
        }
      }
      @_cfg = @_cfg.deep_merge YAML.load_file('/etc/metrics-capacitor.yaml') if File.exists? '/etc/metrics-capacitor.yaml'
    end

    def redis_url
      "redis://#{self.redis[:host]}:#{self.redis[:port].to_s}/#{self.redis[:db].to_s}"
    end

    def influx_url
      [ self.influx[:ssl] ? 'https://' : 'http://',
        self.influx[:host],
        ':',
        self.influx[:port].to_s,
        self.influx[:path],
        "/write?db=",
        self.influx[:db]
      ].join
    end

    def elastic_url
      [ 'http://',
        self.elastic[:host],
        self.elastic[:port],
        self.elastic[:path],
        '/',
        self.elastic[:index],
      ].join
    end

    def worker_path
      File.expand_path('..', __FILE__) + '/processor.rb'
    end

    def method_missing (name, *args, &block)
      return @_cfg[name.to_sym] if @_cfg[name.to_sym] != nil
      fail(NoMethodError, "unknown configuration section #{name}", caller)
    end

  end
end
