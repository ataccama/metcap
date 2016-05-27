module MetricsCapacitor
  module Model
    class Metrics
      extend Forwardable

      def_delegators :@metrics, :slice, :slice!, :map, :each, :empty?, :length, :<<

      def initialize(data = [])
        @metrics = data.map { |m| Metric.new(m) }
      end

      def proc_by_slices!(n)
        @metrics.each_slice(n) { |s| yield Metrics.new(s) }
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
