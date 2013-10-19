# go-configparser

A simple config file parser library for Go.

## Overview

`go-configparser` provides similar functionality to Python's ConfigParser module and deals specifically with settings files in the style of .gitconfig.

## Config file format

`go-configparser` expects config files in the following format:

---------
    [section heading]
    option1 = value1
    option2 = value2

All other lines are ignored - as are any options not found within a section.

See examples/basic.cfg.

## Interpolation

Options from any section can be referenced by using the following format:

---------
    %<setion>(<option>)

Example:

---------
    [global]
    basedir = /opt/myapp
    logs = %local(basedir)/logs

    [app2]
    bindir = %global(basedir)/bin

See examples/interpolate.cfg

## Installation

```
$ go get -u github.com/walkert/go-configparser
```

## Quickstart

---------
    package main

    import (
        "fmt"
        "github.com/walkert/go-configparser"
        "os"
    )

    func main() {
        // Create a new ConfigData object
        cd, err := configparser.Parse("/etc/config.cfg")
        
        // See if the 'global' section exists, and if it does, print the 'main' option value
        if cd.HasSection("global") {
            value, _ := cd.GetOption("global", "main")
            fmt.Println(value)
        }
    }

See examples/configparser_example.go for more details.
