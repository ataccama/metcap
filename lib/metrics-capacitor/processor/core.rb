module MetricsCapacitor
  module Processor
    class Core
      include MetricsCapacitor::Model

      def initialize(logpipe)
        Config.load!
        @_logger = ::Logger.new(logpipe)
        @_name = self.class.to_s.split('::').last.downcase
        @_logger.progname = @_name
        @_logger.level = log_level
        @_logger.formatter = proc { |severity, datetime, progname, msg| "#{progname}##{Process.pid}|||#{severity}|||#{msg}\n" }
        logger.info "Initializing processor"
        post_init
      end

      # implement in the processor Class
      def post_init
      end

      # implement in the processor Class
      def process
        loop { logger.debug "alive"; sleep 10 }
      end

      # implement in the processor Class
      def shutdown
        exit 1
      end

      def logger
        @_logger
      end

      def start!
        logger.info "Starting processor"
        begin
          process
        rescue StandardError => e
          logger.fatal 'SHUTTING DOWN DUE TO AN EXCEPTION!'
          logger.fatal [e.class, e.message].join(' -> ')
          logger.fatal e.backtrace
          logger.fatal 'Sending SIGINT to the engine'
          Process.kill('INT', Process.ppid)
          shutdown!
        end
        logger.info "Processor finished"
      end

      def shutdown!
        shutdown
      end

      private
      def log_level
        Config.debug ? ::Logger::DEBUG : ::Logger::INFO
      end

    end
  end
end
