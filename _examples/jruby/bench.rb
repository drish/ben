require 'benchmark/ips'

Benchmark.ips do |x|
  x.time   = 20
  x.warmup = 3

  x.report { (1..100).inject(:*) }
end
