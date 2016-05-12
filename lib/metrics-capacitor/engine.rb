require 'elasticsearch'
require 'syslog'
require_relative 'config'
require_relative 'sidekiq'
require_relative 'processor/writer'
require_relative 'processor/aggregator'
require_relative 'processor/listener'


module MetricsCapacitor

  class Engine

    def initialize
      $0 = 'metrics-capacitor (engine)'
      @exit_flag = false
      @pids = []
      Config.load!
    end

    def fork_child(args={}, &block)
      args[:proc_num] ||= 1
      args[:exit_on] ||= %w{INT TERM}
      args[:proc_num].times do |num|
        @pids << Process.fork do
          args[:exit_on].each { |sig| Signal.trap(sig) { shutdown! } }
          $0 = "metrics-capacitor (#{args[:name]})" if args[:name]
          begin
            yield block
          rescue StandardError => e
            $stderr.puts e.message
            $stderr.puts e.backtrace.join("\n") if Config.debug
            sleep 1
            retry
          end
        end
      end
    end

    def wait
      # TODO: unix socket for control and status reporting ;)
      begin
        ::Process.waitall
      rescue Interrupt
        retry
      end
    end

    def run_scrubber!
      Config.scrubber[:processes].times do
        @pids << Process.fork do
          $TESTING = 0
          kiq = Sidekiq::CLI.instance
          kiq.parse(['-c', Config.scrubber[:threads].to_s, '-r', Config.scrubber[:worker_path], '-v'])
          kiq.run
        end
      end
    end

    def run_writer!
      fork_child(name: 'writer', proc_num: Config.writer[:processes]) do
        Processor::Writer.new.run!
      end
    end

    def run_aggregator!
      fork_child(name: 'aggregator') do
        include Processor::Aggregator
      end
    end

    def run_listener!
      fork_child(name: 'listener') do
        include Processor::Listener
      end
    end


  end
end
