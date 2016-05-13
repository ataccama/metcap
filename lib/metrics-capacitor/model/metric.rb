module MetricsCapacitor
  module Model
    class Metric
      extend Forwardable

      def_delegators :@metric, :[], :[]=, :merge, :map

      def initialize(data = {})
        @metric = data if data.class == Hash
        @metric ||= MessagePack.unpack(data)
      end

      def to_elastic
        { index: {
            data: {
              :@name      => name,
              :@timestamp => timestamp(:ms),
              :@tags      => tags,
              :@values    => fields
            }
          }
        }
      end

      def to_redis
        @metric.to_msgpack
      end

      def name
        @metric[:name].to_s
      end

      def tags
        return @metrics[:tags].merge({ capacitor: 'tagged' }) unless @metric[:tags].empty?
        { capacitor: 'untagged' }
      end

      def fields
        @metric[:fields]
      end

      def timestamp(scale = :ms)
        m = case scale
            when :ms
              1000.0
            when :us
              1_000_000.0
            when :ns
              1_000_000_000.0
            else
              1.0
            end
        return (Time.now.to_f * m).to_i.to_s unless @metric[:timestamp]
        (Time.at(@metric[:timestamp]).to_f * m).to_i.to_s
      end
    end
  end
end
