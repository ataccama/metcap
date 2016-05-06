require 'forwardable'

module InfluxCapacitor
  class Metrics
    extend Forwardable

    def_delegators :@metrics, :slice, :slice!, :map, :each

    def initialize(data, format = :marshal)
      case format
      when :marshal
        @metrics = Marshal.load(data).map { |m| InfluxCapacitor::Metric.new(m) }
      when :array
        @metrics = data.map { |m| InfluxCapacitor::Metric.new(m) }
      when :json
        @metrics = JSON.load(data).map { |m| InfluxCapacitor::Metric.new(m) }
      end
    rescue StandardError => e
      return nil
    end

    def proc_by_slices!(n)
      @metrics.each_slice(n) { |sl| yield Metrics.new(sl, :array) }
    end

    def to_influx
      @metrics.map(&:to_influx).join("\n")
    end
  end

  class Metric
    extend Forwardable

    def_delegators :@metric, :[], :[]=, :merge, :map

    def initialize(data)
      @metric = data
    end

    def to_influx
      [name, tags, fields, timestamp].join(' ')
    end

    def name
      @metric[:name].to_s
    end

    def tags
      return { capacitor: 'tagged' }.merge(@metric[:tags]).map { |k, v| "#{k}=#{v}" }.join(',') unless @metric[:tags].empty?
      'capacitor=untagged'
    end

    def fields
      @metric[:fields].map { |k, v| "#{k}=#{v.to_f}" }.join(',')
    end

    def timestamp(scale = :ms)
      multiplicator = case scale
                      when :ms
                        1000.0
                      when :us
                        1_000_000.0
                      when :ns
                        1_000_000_000.0
                      else
                        1.0
                      end

      return (Time.now.to_f * multiplicator).to_i.to_s unless @metric[:timestamp]
      (Time.at(@metric[:timestamp]).to_f * multiplicator).to_i.to_s
    end
  end
end
