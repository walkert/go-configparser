package main

import (
	"fmt"
	"github.com/walkert/go-configparser"
	"os"
)

func main() {
	cp := configparser.NewConfigParser(os.Args[1])

	/* Return a map in the form map[string]map[string]string which
	   contain a map of each option->value pair for each section.
	*/
	fmt.Println("Print the data from GetConfigMap()\n")
	confMap, err := cp.GetConfigMap()
	if err != nil {
		fmt.Println("configMap error:", err)
		os.Exit(1)
	}
	for section, options := range confMap {
		fmt.Println("Section:", section)
		for option, value := range options {
			fmt.Printf("Option: %s = %s\n", option, value)
		}
	}

	/* Return a flat map of option->string values. This requires that there
	   are no duplicate options across each of the sections.
	*/
	fmt.Println("\nPrint the data from GetFlatConfigMap()\n")
	flatMap, err := cp.GetFlatConfigMap()
	if err != nil {
		fmt.Println("flatMap error:", err)
		os.Exit(1)
	}
	for option, value := range flatMap {
		fmt.Printf("Option: %s = %s\n", option, value)
	}

	// Or, interrogate the ConfigParser directly
	fmt.Println("\nCheck if the 'global' section exists and print the 'master' option if it does")
	if cp.HasSection("global") {
		value, _ := cp.GetOption("global", "master")
		fmt.Println("master:", value)
	}
	// Try getting an option from an invalid section
	fmt.Println("\nTry and run GetOption on an invalid section")
	value, err := cp.GetOption("app2", "no option")
	if err != nil {
		fmt.Println("Problem!", err)
	} else {
		fmt.Println("Option ok:", value)
	}

	// Try getting an option from an valid section
	fmt.Println("Try and run GetOption on an invalid option")
	value, err = cp.GetOption("app1", "bindir")
	if err != nil {
		fmt.Println("Problem!", err)
		os.Exit(1)
	}
	fmt.Println("Option ok:", value)
}
