# go-configparser

A simple config file parser library for Go.

## Overview

`go-configparser` provides similar functionality to Python's ConfigParser module and deals specifically with settings files in the style of .gitconfig.

Options can also be registered in a similar manner to the flag library in order to provide an explicit set of config requirements.

## Config file format

`go-configparser` expects config files in the following format:

```plain
    [section heading]
    option1 = value1
    option2 = value2
```


All other lines are ignored - as are any options not found within a section.

See examples/basic.cfg.

## Interpolation

Options from any section can be referenced by using the following format:

```plain
    %<section>(<option>)
```

Example:

```plain
    [global]
    basedir = /opt/myapp
    logs = %local(basedir)/logs

    [app2]
    bindir = %global(basedir)/bin
```

See examples/interpolate.cfg

## Installation

```
$ go get -u github.com/walkert/go-configparser
```

## Quickstart

```go
    package main

    import (
        "fmt"
        "github.com/walkert/go-configparser"
        "os"
    )

    func main() {
        // Create a new ConfigParser
        cp := configparser.NewConfigParser("/etc/config.cfg")

        // See if the 'global' section exists, and if it does, print the 'main' option value
        if cp.HasSection("global") {
            value, _ := cp.GetOption("global", "main")
            fmt.Println(value)
        }
    }
```

To instead register options prior to parsing:

```go
    package main

    import (
        "fmt"
        "github.com/walkert/go-configparser"
        "os"
    )

    func main() {
        // Create a new ConfigParser
        cp := configparser.NewConfigParser("/etc/config.cfg")

        // Register an option called 'debug' in the 'main' section
        // with a default value of 'false' and a required value of 'true'.
        // Required options will cause an error if not present.
        debug := cp.BoolOption("debug", "main", false, true)

        err := cp.Parse()

        if err == nil {
            fmt.Println("Error parsing config file: ", err)
        }

        fmt.Println("Debug has been set to: ", *debug)
    }
```

See examples/configparser_example.go for more details.
