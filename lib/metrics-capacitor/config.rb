require 'sidekiq'
require 'sidekiq/logging'
require 'syslog'
require 'log4r'
require 'log4r/configurator'
require 'log4r/outputter/syslogoutputter'

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
        sidekiq_path: `/bin/which sidekiq`.to_s,
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
      begin
        @_cfg = self._cfg.deep_merge YAML.load_file('/etc/influx-capacitor.yaml')
      rescue StandardError => e
        $stderr.puts "Config file load failed:"
        $stderr.puts e.message
        $stderr.puts "using defaults"
      end
    end

    def sidekiq_client_init!
      Sidekiq.configure_client do |config|
        config.redis = { url: self.redis_url }
      end
    end

    def sidekiq_server_init!
      Sidekiq.configure_server do |config|
        config.redis = { url: self.redis_url }
        Sidekiq::Logging.logger = Log4r::Logger.new 'sidekiq'
        Sidekiq::Logging.logger.outputters = self.syslog ? Log4r::SyslogOutputter.new('sidekiq', ident: 'influxdb-capacitor') : Log4r::Outputter.stdout
        Sidekiq::Logging.logger.level = Log4r::INFO
      end
    end

    def deep_merge
    end

    def redis_url
      "redis://#{self._cfg.redis[:host]}:#{self._cfg.redis[:port].to_s}/#{self._cfg.redis[:db].to_s}"
    end

    def influx_url
      [ self._cfg.influx[:ssl] ? 'https://' : 'http://',
        self._cfg.influx[:host],
        ':',
        self._cfg.influx[:port].to_s,
        self._cfg.influx[:path],
        "/write?db=",
        self._cfg.influx[:db]
      ].join
    end

    def es_url
      [ 'http://',
        self._cfg.elastic[:host],
        self._cfg.elastic[:port],
        self._cfg.elastic[:path],
        '/',
        self._cfg.elastic[:index],
      ].join
    end

    def worker_path
      File.expand_path('..', __FILE__) + 'worker.rb'
    end

    def method_missing (name, *args, &block)
      return @_cfg[name.to_sym] if @_cfg[name.to_sym] != nil
      fail(NoMethodError, "unknown configuration section #{name}", caller)
    end

  end
end
