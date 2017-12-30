## Running on Hyper.sh

Before running `ben`, make sure you set:

```
$ export HYPER_ACCESSKEY="your access key"
$ export HYPER_SECRETKEY="your secret key"
$ export HYPER_REGION="us-west-1" // OPTIONAL will default to us-west-1 if not set
```

Then choose a hyper machine type:

```json
{
    "environments": 
    [
        {
            "runtime": "golang",
            "machine": "hyper-s4"
        }
    ]
}
```


Fire !

```
$ ben
```
