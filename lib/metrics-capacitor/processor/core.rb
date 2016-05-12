module MetricsCapacitor
  module Processor
    class Core
      require 'timeout'

      include MetricsCapacitor::Model

      def initialize(logpipe)
        Config.load!
        @_logger = ::Logger.new(logpipe)
        @_name = self.class.to_s.split('::').last.downcase
        @_logger.progname = @_name
        @_logger.formatter = proc { |severity, datetime, progname, msg| "#{progname}##{Process.pid}|||#{severity}|||#{msg}\n" }
        logger.info "Initializing processor"
        post_init
      end

      def post_init
        # implement in the processor Class
      end

      def process
        loop { logger.info "alive"; sleep 10 }
      end

      def shutdown
        exit 1
      end

      def logger
        @_logger
      end

      def start!
        begin
          process
        rescue StandardError => e
          logger.error e.message
          logger.error e.backtrace
          Process.kill('INT', Process.ppid)
          shutdown!
        end
      end

      def shutdown!
        shutdown
      end
    end
  end
end
