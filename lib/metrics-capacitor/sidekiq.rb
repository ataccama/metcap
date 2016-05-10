require 'sidekiq'
require 'sidekiq/logging'
require 'syslog'
require 'log4r'
require 'log4r/configurator'
require 'log4r/outputter/syslogoutputter'

module Sidekiq
  module CLI

  end
end

module MetricsCapacitor

  Config.load!

  Sidekiq.configure_server do |config|
    config.redis = { url: Config.redis_url }
    Sidekiq::Logging.logger = Log4r::Logger.new 'sidekiq'
    Sidekiq::Logging.logger.outputters = Config.syslog ? Log4r::SyslogOutputter.new('sidekiq', ident: 'metrics-capacitor') : Log4r::Outputter.stdout
    Sidekiq::Logging.logger.level = Log4r::INFO
  end
  Sidekiq.configure_client do |config|
    config.redis = { url: Config.redis_url }
  end

end
