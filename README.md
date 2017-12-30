<p align="center">
  <img src="https://rawgit.com/drish/ben/master/assets/ben.png" height="180" />
  <h3 align="center">ben (beta)</h3>
  <p align="center">Your benchmark assistant, written in Go.</p>
  <p align="center">
    <a href="https://travis-ci.org/drish/ben"><img src="https://travis-ci.org/drish/ben.svg?branch=master"></a>
    <a href="https://github.com/drish/ben/blob/master/LICENSE)"><img src="http://img.shields.io/badge/license-MIT-blue.svg"></a>
    <a href="https://goreportcard.com/report/github.com/drish/ben"><img src="https://goreportcard.com/badge/github.com/drish/ben"></a>
  </p>
</p>

---
Ben is a simple tool that helps you running your benchmarks on multiple hardware specs, clouds and runtime versions, so that you can focus on comparing your results.

## Install

With `go get`
```
$ go get https://github.com/drish/ben 
```

## Requirements

- Docker 17.09.1+

## Supported clouds

  * [Hyper.sh](https://hyper.sh)
  * [ECS](https://aws.amazon.com/ecs/) (coming soon.)

## Quick Start

Add a `ben.json` file in the root of your project.

```json
{
  "environments": [
    {
      "runtime": "ruby",
      "version": "2.3",
      "machine": "local",
      "before": ["gem install benchmark-ips"],
      "command": "ruby bench.rb"
    },
    {
      "runtime": "ruby",
      "version": "2.5",
      "machine": "local",
      "before": ["gem install benchmark-ips"],
      "command": "ruby bench.rb"
    }
  ]
}

```


Then, in the root of your project run.

```
$ ben
```

This will tell Ben to run two benchmarks runtimes for ruby: 2.3 and 2.5, on your **local** machine using docker containers.
It will run the commands specified on the `before` field and will create a new docker image with all set for your benchmark.
After all benchmarks are done, you can see the results at `./benchmarks.md`

Checkout the [examples](https://github.com/drish/ben/tree/master/_examples) folder for more examples.

## Running on Hyper.sh

Make sure you set

```
$ export HYPER_ACCESSKEY="your access key"
$ export HYPER_SECRETKEY="your secret key"
$ export HYPER_REGION="us-west-1" // OPTIONAL will default to us-west-1 if not set
```

## ben.json file spec

```
{
  "environments": [
    {
      "runtime": "", // required, ie: golang
      "version": "", // OPTIONAL, default to "latest", ie: 1.3
      "machine": "", // OPTIONAL, default to "local", ie: hyper-s1
      "command": "", // OPTIONAL
      "before": [""] // OPTIONAL
    }
  ]
}
```

Each field is described below.

#### runtime

Base docker image that will be used for benchmarking. 
Example: `golang`, `ruby`, `jruby`
Final docker image to run is composed by `runtime`:`version`

#### version

Base image tag, default to `latest`, ie: `1.8`
Final docker image to run is composed by `runtime`:`version`

#### machine

Machine type, default to `local` which will run your benchmarks on local docker containers.

For running locally, options is: 

  * `local`

For running on hyper.sh cloud, options are: 

  * `hyper-s1` (1 CPU 64MB)
  * `hyper-s2` (1 CPU 124MB)
  * `hyper-s3` (1 CPU 256MB)
  * `hyper-s4` (1 CPU 512MB)
  * `hyper-m1` (1 CPU 1GB)
  * `hyper-m2` (2 CPU 2GB)
  * `hyper-m3` (2 CPU 4GB)
  * `hyper-l1` (4 CPU 4GB)
  * `hyper-l2` (4 CPU 8GB)
  * `hyper-l3` (8 CPU 16GB)

#### command

Benchmark command to run.
If not set, a default command is set based on your runtime, right now a default command is only set for runtime `golang`.

runtime |     command      |
--------|------------------|
golang  | go test -bench=. |

#### before

Commands to run before your benchmark command.

Example, if you your benchmark scripts depend on `quemu-img` library to run , you can set your `before` field like:


```json
"before": ["apt-get install qemu-img"]
```

## LICENSE

[MIT](https://github.com/drish/ben/blob/master/LICENSE)
