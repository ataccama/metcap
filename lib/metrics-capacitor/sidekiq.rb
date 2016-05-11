require 'sidekiq'
require 'sidekiq/cli'
require 'sidekiq/logging'
require 'syslog'
require 'log4r'
require 'log4r/configurator'
require 'log4r/outputter/syslogoutputter'

module Sidekiq
  class CLI

    PROCTITLES[0] = proc { 'metrics-capacitor'.freeze }
    PROCTITLES[1] = proc { '(scrubber)'.freeze }

    def run(*args)
      @code = nil
      setup_options(args)
      initialize_logger
      validate!
      write_pid
      boot_system

      self_read, self_write = IO.pipe

      %w(INT USR1 USR2 TTIN).each do |sig|
        begin
          trap sig do
            self_write.puts(sig)
          end
        rescue ArgumentError
          puts "Signal #{sig} not supported"
        end
      end

      ver = Sidekiq.redis_info['redis_version']
      raise "You are using Redis v#{ver}, Sidekiq requires Redis v2.8.0 or greater" if ver < '2.8'
      fire_event(:startup)
      logger.debug { "Client Middleware: #{Sidekiq.client_middleware.map(&:klass).join(', ')}" }
      logger.debug { "Server Middleware: #{Sidekiq.server_middleware.map(&:klass).join(', ')}" }
      require 'sidekiq/launcher'
      @launcher = Sidekiq::Launcher.new(options)
      begin
        launcher.run
        while readable_io = IO.select([self_read])
          signal = readable_io.first[0].gets.strip
          handle_signal(signal)
        end
      rescue Interrupt
        logger.info 'Shutting down'
        launcher.stop
        logger.info "Bye!"
        exit(0)
      end
    end
  end
end

module MetricsCapacitor

  Config.load!

  Sidekiq.configure_server do |config|
    config.redis = { url: Config.redis_url }
    Sidekiq::Logging.logger = Log4r::Logger.new 'sidekiq'
    Sidekiq::Logging.logger.outputters = Config.syslog ? Log4r::SyslogOutputter.new('sidekiq', ident: 'metrics-capacitor') : Log4r::Outputter.stdout
    Sidekiq::Logging.logger.level = Log4r::INFO
  end
  Sidekiq.configure_client do |config|
    config.redis = { url: Config.redis_url }
  end

end
