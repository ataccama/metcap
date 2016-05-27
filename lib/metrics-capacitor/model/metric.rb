module MetricsCapacitor
  module Model
    class Metric
      extend Forwardable

      def_delegators :@metric, :[], :[]=, :merge, :map

      def initialize(data = {})
        @metric = data if data.class == Hash
        @metric ||= JSON.parse(data, symbolize_names: true)
      end

      def to_elastic
        { index: {
            data: {
              :@name      => name,
              :@timestamp => timestamp(:ms),
              :@tags      => tags,
              :@values    => values
            }
          }
        }
      end

      def to_redis
        @metric.to_json
      end

      def name
        @metric[:name].to_s
      end

      def tags
        return @metric[:tags] if ( @metric[:tags] || @metric[:tags].empty? )
        { capacitor: 'untagged' }
      end

      def values
        case @metric[:values]
        when Hash
          return @metric[:values]
        when Integer
          return { value: @metric[:values].to_f }
        when Float
          return { value: @metric[:values] }
        else
          return { value: 0.0 }
        end
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
