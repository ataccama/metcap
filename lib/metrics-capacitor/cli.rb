require 'thor'
require_relative 'config'

module MetricsCapacitor
  class CLI < Thor

    @exit = false

    $0 = 'metrics-capacitor'
    Config.load!
    package_name 'Metrics Capacitor'

    desc 'daemon', 'Run the workers'
    option :concurrency, type: :numeric, default: Config.concurrency, aliases: :c
    def daemon
      cmd = [ Config.sidekiq_path, '-c', options[:concurrency].to_s, '-r', Config.worker_path ]
      system *cmd
    end

    desc 'status', 'Report the state'
    def status
      Config.sidekiq_init_client!
      # TODO ...
    end
  end

  def fork_child(*args, &block)
    Process.fork do
      %w(INT TERM KILL).each { |sig| Signal.trap(sig) { @exit = true } }
      yield block
    end
  end

  def fork_sidekiq(*args, &block)
    Process.fork do
      %w(INT TERM KILL).each do |sig| Signal.trap(sig) { exit 0 } }
    end
  end
end
