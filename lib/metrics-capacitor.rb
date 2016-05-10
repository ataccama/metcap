require_relative 'metrics-capacitor/config'
require_relative 'metrics-capacitor/sidekiq'

module MetricsCapacitor
  require_relative 'metrics-capacitor/model'
  require_relative 'metrics-capacitor/processors/scrubber'
  require_relative 'metrics-capacitor/processors/' + Config.storage_engine.to_s

  # for metrics scrubbing
  # implemented as a Sidekiq job
  #
  class Anode
    include Sidekiq::Worker
    include Processors::Scrubber
    sidekiq_options retry: true
    sidekiq_options queue: 'metrics'
  end

  # metrics data writer
  #
  class Cathode
    include Kernel.const_get "Processors::#{Config.storage_engine.to_s.capitalize}"
  end

  # proof of concept worker
  # implemented as a Sidekiq job
  #
  class Worker
    include Sidekiq::Worker
    include Kernel.const_get "Processors::#{Config.storage_engine.to_s.capitalize}"
    sidekiq_options retry: true
    sidekiq_options queue: 'metrics'
  end
end
