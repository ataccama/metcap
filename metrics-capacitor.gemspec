Gem::Specification.new do |s|
  s.name                  = 'metrics-capacitor'
  s.version               = '0.0.1'
  s.date                  = Time.now.strftime('%Y-%m-%d')
  s.summary               = "Metrics Capacitor"
  s.description           = "Metrics igenstion service with ElasticSearch as storage"
  s.authors               = ["Radek 'blufor' Slavicinsky"]
  s.email                 = 'radek@blufor.cz'
  s.files                 = Dir['lib/**/*.rb']
  s.executables           = Dir['bin/*'].map(){ |f| f.split('/').last }
  s.homepage              = 'https://github.com/metrics-capacitor/metrics-capacitor'
  s.license               = 'GPLv3'
  s.required_ruby_version = '>= 2.0.0'
  s.add_runtime_dependency 'thor', '~> 0.19', '>= 0.19.1'
  s.add_runtime_dependency 'msgpack', '~> 0.7', '>= 0.7.6'
end
