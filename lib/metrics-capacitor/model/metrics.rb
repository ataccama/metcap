module MetricsCapacitor
  module Model
    class Metrics
      extend Forwardable

      def_delegators :@metrics, :slice, :slice!, :map, :each, :empty?, :length

      def initialize(data)
        @metrics = data.map { |m| MetricsCapacitor::Metric.new(m) } if data.class == Array
        @metrics ||= MessagePack.unpack(data).map { |m| MetricsCapacitor::Metric.new(m) }
      rescue StandardError => e
        $stderr.puts e.message
        return nil
      end

      def proc_by_slices!(n)
        @metrics.each_slice(n) { |s| yield Metrics.new(s, :array) }
      end

      def to_elastic
        @metrics.map(&:to_elastic)
      end

      def to_redis
        @metrics.map(&:to_redis)
      end
    end
  end
end
