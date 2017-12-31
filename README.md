<p align="center">
  <img src="https://rawgit.com/drish/ben/master/assets/ben.png" height="180" />
  <h3 align="center">Ben (beta)</h3>
  <p align="center">Your benchmark assistant, written in Go.</p>
  <p align="center">
    <a href="https://travis-ci.org/drish/ben"><img src="https://travis-ci.org/drish/ben.svg?branch=master"></a>
    <a href="https://github.com/drish/ben/blob/master/LICENSE)"><img src="http://img.shields.io/badge/license-MIT-blue.svg"></a>
    <a href="https://goreportcard.com/report/github.com/drish/ben"><img src="https://goreportcard.com/badge/github.com/drish/ben"></a>
  </p>
</p>

---

Ben is a simple tool that helps you run your benchmarks on multiple hardware specs, clouds and runtime versions.

## Install

With `go get`
```
$ go get https://github.com/drish/ben/cmd/ben
```

or with `curl`

```
curl -sf https://raw.githubusercontent.com/drish/ben/master/install.sh | sh
```

## Requirements

- Docker 17.03.0-ce+

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

After all benchmarks are done, a [benchmarks.md](https://github.com/drish/ben/tree/master/_examples/go/local/benchmarks.md) file will be generated.

Checkout [examples](https://github.com/drish/ben/tree/master/_examples) folder for more.

---

<p align="center">
  <img src="https://rawgit.com/drish/ben/master/assets/demo.gif"/>
</p>

---

### More docs

  * [Running on hyper.sh](https://github.com/drish/ben/blob/master/docs/running-on-hyper.md)
  * [ben.json file spec](https://github.com/drish/ben/blob/master/docs/ben-json-spec.md)

## License

MIT © [Carlos Derich](https://dri.sh)
