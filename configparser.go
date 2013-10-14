package configparser

import (
    "bufio"
    "errors"
    "fmt"
    "regexp"
    "os"
)

type Option struct {
    Val string
}

type SectionData struct {
    Options map[string]Option
}

func HasOption (c *ConfigData, section, option string) bool {
    if _, ok := c.Sections[section].Options[option] ; ok {
        return true
    } else {
        return false
    }
}

func AddOption (c *ConfigData, section, option, value string) {
    newOption := Option{value}
    c.Sections[section].Options[option] = newOption
}

type ConfigData struct {
    Sections map[string]SectionData
    Interpolate map[string]map[string]string
}

func NewConfigData () ConfigData {
    sections := make(map[string]SectionData)
    interpolate := make(map[string]map[string]string)
    return ConfigData{sections, interpolate}
}

func (c *ConfigData) HasSection (section string) bool {
    if _, ok := c.Sections[section] ; ok {
        return true
    } else {
        return false
    }
}

func (c *ConfigData) HasSecOpt (section, option string) bool {
    if HasOption(c, section, option) {
        return true
    } else {
        return false
    }
}

func (c *ConfigData) AddSecOpt (section, key, value string) {
    AddOption(c, section, key, value)
}

func (c *ConfigData) AddSection (section string) {
    sd := SectionData{make(map[string]Option)}
    sd.Options = make(map[string]Option)
    c.Sections[section] = sd
}

//func (c *ConfigData) Parse (fname string) (ConfigData, error) {
func Parse (fname string) (ConfigData, error) {
    fd, err := os.Open(fname)
    var returnedError error
    if err != nil {
        returnedError = err
    }
    defer fd.Close()
    var section string
    cd := NewConfigData()
    secMatch := regexp.MustCompile(`^\[(.*)\]$`)
    confPair := regexp.MustCompile(`(\w*)\s+=\s+(\w*)`)
    scanner := bufio.NewScanner(fd)
    for scanner.Scan() {
        line := scanner.Text()
        switch {
        case secMatch.MatchString(line):
            section = secMatch.FindStringSubmatch(line)[1]
            if cd.HasSection(section) {
                etext := fmt.Sprintf("Duplicate section found: %s", section)
                returnedError = errors.New(etext)
                break
            } else {
                cd.AddSection(section)
            }
        case confPair.MatchString(line):
            pair := confPair.FindStringSubmatch(line)
            key := pair[1]
            val := pair[2]
            if cd.HasSecOpt(section, key) {
                etext := fmt.Sprintf("Duplicate option %s found in section %s.", key, section)
                returnedError = errors.New(etext)
                break
            } else {
                cd.AddSecOpt(section, key, val)
            }
        }
    }
    return cd, returnedError
}

func (c *ConfigData) Printer () {
    for section, secdata := range(c.Sections) {
        fmt.Println("Section:", section)
        for option, value := range(secdata.Options) {
            fmt.Printf("Option: %s == %s\n", option, value.Val)
        }
    }
}
