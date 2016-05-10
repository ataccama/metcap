module MetricsCapacitor
  class Process

    def initialize
      @exit_flag = false
      @pids = [ ::Process.pid ]
    end

    $0 = 'metrics-capacitor'

    class << self

      def fork_child(*args, &block)
        ::Process.fork do
          %w(INT TERM KILL).each { |sig| Signal.trap(sig) { @exit = true } }
          $0 = "metrics-capacitor [#{args[:name]}]" if args[:name]
          yield block
        end
      end

      def run_sidekiq(*args)
        ::Process.fork do
          %w(INT TERM KILL).each { |sig| Signal.trap(sig) { exit 0 } }
          $0 = "metrics-capacitor [sidekiq]"
        end
      end

    end

  end
end
