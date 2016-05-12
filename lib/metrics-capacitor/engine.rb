require 'elasticsearch'
require 'syslog'
require 'syslog/logger'
require_relative 'config'
require_relative 'sidekiq'
require_relative 'model'
require_relative 'logger'
require_relative 'processor/core'
require_relative 'processor/writer'
require_relative 'processor/aggregator'
require_relative 'processor/listener'


module MetricsCapacitor

  class Engine

    def initialize
      $0 = 'metrics-capacitor (engine)'
      Config.load!
      @exit_flag = false
      @pids = []
      %w(TERM INT).each do |sig|
        Signal.trap(sig) do
          @pids.each { |pid| Process.kill(sig, pid) rescue true }
          Process.waitall
          terminate_loggers
        end
      end
      @logpipe = {}
      @logger = ::Logger.new(STDOUT)
      @logger.level = log_level
      @logger.formatter = proc { |severity, datetime, progname, msg| [datetime.to_s, progname, severity, "#{msg}\n"].join(" ") }
      @logger_threads = []
      @logger_semaphore = Mutex.new
      # Logger.init!
      log :info, "Initialized :-)"
    end

    def fork_processor(args = {})
      log :debug, "Spawning #{args[:name]}"
      args[:proc_num] ||= 1
      args[:exit_on] ||= %w{INT TERM}
      args[:proc_num].times do |num|
        @logpipe["#{args[:name]}_#{num}".to_sym], logpipe = IO.pipe
        @pids << Process.fork do
          $0 = "metrics-capacitor (#{args[:name]})" if args[:name]
          @logpipe["#{args[:name]}_#{num}".to_sym].close
          remove_instance_variable(:@logpipe)
          p = Kernel.const_get("MetricsCapacitor::Processor::#{args[:name].capitalize}").new(logpipe)
          args[:exit_on].each { |sig| Signal.trap(sig) { p.shutdown! } }
          p.start!
        end
        log :debug, "Processor #{args[:name]} spawned as PID #{@pids.last.to_s}"
        logpipe.close
        sleep 1
      end
    end

    def fork_scrubber
      log :debug, "Spawning scrubbers"
      Config.scrubber[:processes].times do |num|
        @logpipe["scrubber_#{num}".to_sym], logpipe = IO.pipe
        @pids << Process.fork do
          @logpipe["scrubber_#{num}".to_sym].close
          remove_instance_variable(:@logpipe)
          Sidekiq.configure_server do |config|
            Sidekiq::Logging.logger = ::Logger.new(logpipe)
            Sidekiq::Logging.logger.level = log_level
            Sidekiq::Logging.logger.progname = "scrubber"
            Sidekiq::Logging.logger.formatter = proc { |severity, datetime, progname, msg| "#{progname}##{Process.pid}|||#{severity}|||#{msg}\n" }
            config.redis = { url: Config.redis[:url] }
          end
          Sidekiq.configure_client do |config|
            config.redis = { url: Config.redis[:url] }
          end
          $TESTING = 0
          kiq = Sidekiq::CLI.instance
          kiq.parse(['-c', Config.scrubber[:threads].to_s, '-r', Config.scrubber[:worker_path]])
          kiq.run
        end
        logpipe.close
      end
    end

    def log(severity = :info, msg)
      s = Kernel.const_get("Logger::#{severity.to_s.upcase}")
      @logger_semaphore.synchronize do
        @logger.log s, msg, 'engine'
      end
    end

    def log_level
      Config.debug ? ::Logger::DEBUG : ::Logger::INFO
    end

    def spawn_loggers
      @logpipe.each do |name, pipe|
        @logger_threads << Thread.new do
          Thread.current[:name] = "logger-#{name}"
          while msg = pipe.gets
            (progname,severity,message) = msg.split('|||')
            # $stderr.puts "#{progname} #{severity} #{message}"
            @logger_semaphore.synchronize do
              @logger.log Kernel.const_get("Logger::#{severity}"), message.chomp, progname
            end
          end
        end
      end
    end

    def terminate_loggers
      @logger_threads.each { |t| t.join }
    end

    def run!
      log :info, 'Spawning processes'
      fork_scrubber
      fork_processor name: 'writer', proc_num: Config.writer[:processes]
      fork_processor name: 'aggregator'
      fork_processor name: 'listener'
      spawn_loggers
      # TODO: unix socket for control and status reporting ;)
      begin
        ::Process.waitall
      rescue Interrupt
        retry
      end
      log :info, "Terminating loggers"
      terminate_loggers
      log :warn, "Engine is shutting down!"
    end


  end
end
