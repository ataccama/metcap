require 'excon'
require 'influx-capacitor/config'
require 'influx-capacitor/metrics'

InfluxCapacitor::Config.load!
InfluxCapacitor::Config.sidekiq_server_init!
InfluxCapacitor::Config.sidekiq_client_init!

module InfluxCapacitor
  class Worker < Sidekiq::Worker
    sidekiq_options retry: true
    sidekiq_options queue: 'metrics'

    INFLUX = ConnectioPool.new(size: Config.influx[:connections]) do
      Excon.new Config.influx_url,
        persistent: true,
        expects: [204],
        method: :post,
        read_timeout: Config.influx[:timeout],
        write_timeout: Config.influx[:timeout],
        connect_timeout: Config.influx[:timeout],
        idempotent: true,
        retry_limit: Config.influx[:retry],
        tcp_nodelay: true
    end

    def process *args
      # parse data
      begin
        influx_data = Metrics.new(args).to_influx
      rescue StandardError => e
      end

      # write data
      begin
        INFLUX.with { |i| i.post body: influx_data }
      rescue StandardError => e
      end
    end

  end
end
