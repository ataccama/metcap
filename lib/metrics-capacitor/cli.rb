require 'thor'
require_relative 'config'

module MetricsCapacitor
  class CLI < Thor
    include MetricsCapacitor::Process

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
end
