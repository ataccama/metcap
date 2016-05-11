require_relative 'config'
require_relative 'sidekiq'

module MetricsCapacitor
  class Engine

    def initialize
      $0 = 'metrics-capacitor'
      @exit_flag = false
      @pids = [ ::Process.pid ]
      Config.load!
    end


    def fork_child(*args, &block)
      @pids << Process.fork do
        args[:exit_on].each { |sig| Signal.trap(sig) { @exit = true } }
        $0 = "metrics-capacitor (#{args[:name]})" if args[:name]
        loop do
          _t_start = Time.now.to_i
          yield block
          _sleep = args[:interval] - _t_start + Time.now.to_i
          sleep _sleep unless _sleep.negative?
        end
      end
    end

    def wait
      # TODO process state checking with SIG handling
      ::Process.waitall
    end

    def fork_sidekiq(*args)
      args[:processes] ||= 4
      loop(args[:processes]) do
        @pids << Process.fork do
          Sidekiq::CLI.instance.run(['-c', args[:concurrency].to_s, '-r', args[:worker]])
        end
      end
    end

    def run_scrubber!
      fork_sidekiq processes: Config.scrubber[:processes], concurrency: Config.scrubber[:concurrency], worker: Config.scrubber[:worker_path]
    end


    def run_writer!
      fork_child(name: 'writer', exit_on: ['INT']) do
        MetricsCapacitor::Processor.new
    end

    def run_aggregator!
      fork_child(name: 'aggregator')  do
        MetricsCapacitor::
      end
    end

  end
end
