# falcon

## In development

Resumable download accelerator written in Go-lang

# Installation

Install dependencies with [dep](https://github.com/golang/dep#setup)

```
dep ensure
```

Manual build

```
./bin/buid
```

# Usage

```
falcon [cmd]
```

Here is a list of available commands:

```
Usage:
  falcon [command]

Available Commands:
  get         Download the given url
  help        Help about any command
  resume      Resume unfinished task
  tasks       Listing all unfinished tasks

Flags:
  -h, --help   help for falcon

Use "falcon [command] --help" for more information about a command.
```


```
./build/falcon get --help
```

```
Download the given url

Usage:
  falcon get [url] [flags]

Flags:
  -c, --connection int   The number of connections (default 4)
  -h, --help             help for get
```
# LICENSE

[MIT](LICENSE)