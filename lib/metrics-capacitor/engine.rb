require_relative 'config'
require_relative 'sidekiq'

module MetricsCapacitor
  class Engine
    require_relative 'metrics-capacitor/model'
    require_relative 'metrics-capacitor/processors/scrubber'
    require_relative 'metrics-capacitor/processors/writer'
    require_relative 'metrics-capacitor/processors/aggregator'

    def initialize
      $0 = 'metrics-capacitor'
      @exit_flag = false
      @pids = [ ::Process.pid ]
      Config.load!
    end

    def fork_child(*args, &block)
      @pids << Process.fork do
        args[:loop] ||= false
        args[:interval] ||= 0
        args[:exit_on] ||= %w{INT TERM}
        args[:exit_on].each { |sig| Signal.trap(sig) { raise Interrupt } }
        $0 = "metrics-capacitor (#{args[:name]})" if args[:name]
        _p = lambda { |&_block|
          begin
            _t_start = Time.now.to_i
            _block.call
            _sleep = args[:interval] - _t_start + Time.now.to_i
            sleep _sleep unless _sleep.negative?
          rescue Interrupt
            exit
          end
        }
        if args[:loop]
          loop(&_p.call do
            yield block
          end)
        else
          _p.call { yield block }
        end
      end
    end

    def wait
      # TODO process state checking with SIG handling
      ::Process.waitall
    end

    def fork_sidekiq(*args)
      args[:processes] ||= 1
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
      loop(Config.writer[:processes]) do
        fork_child(name: 'writer', exit_on: ['INT']) do
          MetricsCapacitor::Processor.new
        end
      end
    end

    def run_aggregator!
      fork_child(name: 'aggregator')  do
        MetricsCapacitor::Aggregator.new
      end
    end

  end
end
