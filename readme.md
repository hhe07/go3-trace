# go3-view

## usage
run the program via
```sh
    ./o3view [options] [input file path]
```

i.e.
```sh
    ./o3view o3.txt
```

learn more about the options through
```sh
    ./o3view --help
```

and through viewing ``main.go``.

## notes / warnings

many of the options are similar to that of the ``util/o3-pipeview.py`` script, except that there's no option to directly print to stdout.

in my testing, the program uses upwards of ~50% of CPU when running.

## pre-built binaries
check ``bin/``. only x86_64 linux has been tested. 

try ``go tool dist list`` for supported (cross) compilation architectures.

## compilation

in the root directory of your local copy,
```sh
    go mod download
    go build .
```
