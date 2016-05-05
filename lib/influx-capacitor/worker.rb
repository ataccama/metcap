require 'net/http'
require 'influx-capacitor/config'
require 'influx-capacitor/metrics'

InfluxCapacitor::Config.load!
InfluxCapacitor::Config.sidekiq_server_init!
InfluxCapacitor::Config.sidekiq_client_init!

module InfluxCapacitor
  class Worker < Sidekiq::Worker
    sidekiq_options retry: true
    sidekiq_options queue: 'metrics'

    $influx_conn = ConnectioPool.new {  } # FIXME: finish

    def process *args
      influx_data = InfluxCapacitor::Metrics.new(args).to_influx
      # $influx_conn.request =
    end

  end
end
