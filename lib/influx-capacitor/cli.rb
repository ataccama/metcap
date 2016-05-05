require 'thor'

module InfluxCapacitor
  class CLI < Thor
    Config.load!
    package_name 'Influx Capacitor'

    desc 'run', 'Run the workers'
    option :concurrency, type: :numeric, default: Config.concurrency, aliases: :c
    def run
      Process.setproctitle 'influxdb-capacitor'
      system Config.sidekiq_path, '-c', options[:concurrency], '-r', Config.worker_path
    end

    desc 'status', 'Report the state'
    def status
      Config.sidekiq_init_client!
    end
end
