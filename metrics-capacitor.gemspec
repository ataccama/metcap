Gem::Specification.new do |s|
  s.name                  = 'metrics-capacitor'
  s.version               = '0.0.1'
  s.date                  = Time.now.strftime('%Y-%m-%d')
  s.summary               = "Metrics Capacitor"
  s.description           = "Sidekiq worker for crunching metrics data (mainly gathered by Sensu) into Metrics DB"
  s.authors               = ["Radek 'blufor' Slavicinsky"]
  s.email                 = 'radek.slavicinsky@gmail.com'
  s.files                 = Dir['lib/**/*.rb']
  s.executables           = Dir['bin/*'].map(){ |f| f.split('/').last }
  s.homepage              = 'https://github.com/prozeta/metrics-capacitor'
  s.license               = 'GPLv3'
  s.required_ruby_version = '>= 1.9.3'
  s.add_runtime_dependency 'sidekiq', '~> 4.1', '>= 4.1.2'
  s.add_runtime_dependency 'thor', '~> 0.19', '>= 0.19.1'
end
