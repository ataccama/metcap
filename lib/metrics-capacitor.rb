require_relative 'metrics-capacitor/config'
require_relative 'metrics-capacitor/sidekiq'

module MetricsCapacitor
  class Scrubber
    include Sidekiq::Worker
    include Processors::Scrubber
    sidekiq_options retry: true
    sidekiq_options queue: 'scrubber'
  end

  # metrics data writer
  #
  class Writer
    include Processors::Writer
  end

  # metrics data aggregator
  #
  class Aggregator
    include Processors::Aggregator
  end

  # proof of concept worker
  # implemented as a Sidekiq job
  #
  class Worker
    include Sidekiq::Worker
    include Processors::Elastic
    sidekiq_options retry: true
    sidekiq_options queue: 'metrics'
  end
end
