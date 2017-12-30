## ben.json file spec

```
{
  "environments": [
    {
      "runtime": "", // REQUIRED, ie: golang
      "version": "", // OPTIONAL, default to "latest", ie: 1.3
      "machine": "", // OPTIONAL, default to "local", ie: hyper-s1
      "command": "", // OPTIONAL
      "before": [""] // OPTIONAL
    }
  ]
}
```

Each field is described below.

### runtime

Base docker image that will be used for benchmarking. 
Example: `golang`, `ruby`, `jruby`, `node`

Final docker image to run is composed by `runtime`:`version`

### version

This is translated to a docker image tag, default to `latest`, if not set.
Example: `1.8`.

### machine

Machine type, default to `local` which will run your benchmarks on local docker containers.

For running locally, options is: 

  * `local`

For running on **hyper.sh cloud**, options are: 

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

### command

Benchmark command to run.
If not set, a default command is set based on your runtime, right now a default command is only set for runtime `golang`.

runtime |     command      |
--------|------------------|
golang  | go test -bench=. |

### before

Commands to run before your benchmark command.

Example, if you your benchmark scripts depend on `npm install`, you can set your `before` field like:


```json
"before": ["npm install"]
```
