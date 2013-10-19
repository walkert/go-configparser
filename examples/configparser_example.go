package main

import (
    "fmt"
    "github.com/walkert/go-configparser"
    "os"
)

func main() {
    /* Create a new ConfigData object. An error will be produced if 
       the config file doesn't fit the expected style.
    */

    cd, err := configparser.Parse(os.Args[1])
    if err != nil {
        fmt.Println("Got err:", err)
        os.Exit(1)
    }

    /* Return a map in the form map[string]map[string]string which
       contain a map of each option->value pair for each section.
    */
    fmt.Println("Print the data from GetConfigMap()\n")
    confMap := cd.GetConfigMap()
    for section, options := range(confMap) {
        fmt.Println("Section:", section)
        for option, value := range(options) {
            fmt.Printf("Option: %s = %s\n", option, value)
        }
    }

    /* Return a flat map of option->string values. This requires that there
       are no duplicate options across each of the sections.
    */
    fmt.Println("\nPrint the data from GetFlatConfigMap()\n")
    flatMap, err := cd.GetFlatConfigMap()
    if err != nil {
        fmt.Println("flatMap error:", err)
        os.Exit(1)
    }
    for option, value := range(flatMap) {
        fmt.Printf("Option: %s = %s\n", option, value)
    }

    // Or, interrogate the ConfigData directly
    fmt.Println("\nCheck if the 'global' section exists and print the 'master' option if it does")
    if cd.HasSection("global") {
        value, _ := cd.GetOption("global", "master")
        fmt.Println("master:", value)
    }
    // Try getting an option from an invalid section
    fmt.Println("\nTry and run GetOption on an invalid section")
    value, err := cd.GetOption("app2", "no option")
    if err != nil {
	    fmt.Println("Problem!", err)
    } else {
	    fmt.Println("Option ok:", value)
    }

    // Try getting an option from an valid section
    fmt.Println("Try and run GetOption on an invalid option")
    value, err = cd.GetOption("app1", "bindir")
    if err != nil {
	    fmt.Println("Problem!", err)
	    os.Exit(1)
    }
    fmt.Println("Option ok:", value)
}
