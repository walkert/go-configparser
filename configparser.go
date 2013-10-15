package configparser

import (
    "bufio"
    "errors"
    "fmt"
    "regexp"
    "os"
)

type InterpolateObj struct {
    Value string
    Global bool
}

type SectionData struct {
    Options map[string]string
    Interpolate map[string]InterpolateObj
}

func HasOption (c *ConfigData, section, option string) bool {
    if _, ok := c.Sections[section].Options[option] ; ok {
        return true
    } else {
        return false
    }
}

func AddOption (c *ConfigData, section, option, value string) {
    c.Sections[section].Options[option] = value
}

func AddInter (c *ConfigData, section, option string, idata []string) error {
    var retval error = nil
    var global bool
    designator := idata[0]
    value := idata[1]
    switch {
        case designator == "g":
            global = true
        case designator == "l":
            global = false
        default:
            etext := fmt.Sprintf("Invalid interpolation designator: %s", designator)
            return errors.New(etext)
    }
    inter := InterpolateObj{value, global}
    c.Sections[section].Interpolate[option] = inter
    return retval
}

type ConfigData struct {
    Sections map[string]SectionData
    Regexps map[string]*regexp.Regexp
}

func NewConfigData () ConfigData {
    secdata := make(map[string]SectionData)
    regexps := make(map[string]*regexp.Regexp)
    regexps["inter"] = regexp.MustCompile(`^%([a-z])\((.*)\)`)
    regexps["secMatch"] = regexp.MustCompile(`^\[(.*)\]$`)
    regexps["confPair"] = regexp.MustCompile(`^(\w*)\s+=\s+(.*)`)
    return ConfigData{secdata, regexps}
}

func (c *ConfigData) HasSection (section string) bool {
    if _, ok := c.Sections[section] ; ok {
        return true
    } else {
        return false
    }
}

func (c *ConfigData) Regexp (regexp string) *regexp.Regexp {
    return c.Regexps[regexp]
}

func (c *ConfigData) GetSecOpt (section, option string) string {
    return c.Sections[section].Options[option]
}

func (c *ConfigData) SetSecOpt (section, option, value string) {
    c.Sections[section].Options[option] = value
}

func (c *ConfigData) HasSecOpt (section, option string) bool {
    if HasOption(c, section, option) {
        return true
    } else {
        return false
    }
}

func (c *ConfigData) AddSecOpt (section, key, value string) error {
    var retval error = nil
    AddOption(c, section, key, value)
    igroup := c.Regexp("inter").FindStringSubmatch(value)
    if len(igroup) == 3 {
        retval = AddInter(c, section, key, igroup[1:])
    }
    return retval
}

func (c *ConfigData) AddSection (section string) {
    options := make(map[string]string)
    interpolate := make(map[string]InterpolateObj)
    sd := SectionData{options, interpolate}
    c.Sections[section] = sd
}

func (c *ConfigData) GetSections () []string {
    optionList := make([]string, len(c.Sections))
    for key, _ := range(c.Sections) {
        optionList = append(optionList, key)
    }
    return optionList
}

func (c *ConfigData) Interpolation () error {
    var retval error = nil
    var refsection string
    for _, section := range(c.GetSections()) {
        if len(c.Sections[section].Interpolate) > 0 {
            for key, irefval := range(c.Sections[section].Interpolate) {
                if irefval.Global {
                    refsection = "global"
                } else {
                    refsection = section
                }
                if c.HasSection(refsection) {
                    if c.HasSecOpt(refsection, key) {
                        c.Interpolate(section, key, refsection, irefval.Value)
                    } else { 
                        etext := fmt.Sprintf("Cannot interpolate %s, %s section does not contain key %s.", refsection, key)
                        return errors.New(etext)
                    }
                } else {
                    etext := fmt.Sprintf("Cannot interpolate %s, %s section not defined.", key, refsection)
                    return errors.New(etext)
                }
            }
        }
    }
    return retval
}

func (c *ConfigData) Interpolate (section, key, refsection, irefkey string) {
    current := c.GetSecOpt(section, key)
    ival := c.GetSecOpt(refsection, irefkey)
    replacement := c.Regexp("inter").ReplaceAllLiteralString(current, ival)
    c.SetSecOpt(section, key, replacement)
}
    
func Parse (fname string) (ConfigData, error) {
    var returnedError error = nil
    blank := NewConfigData()
    cd := NewConfigData()
    fd, err := os.Open(fname)
    if err != nil {
        returnedError = err
        return cd, returnedError
    }
    defer fd.Close()
    var section string
    scanner := bufio.NewScanner(fd)
    for scanner.Scan() {
        line := scanner.Text()
        switch {
        case cd.Regexp("secMatch").MatchString(line):
            section = cd.Regexp("secMatch").FindStringSubmatch(line)[1]
            if cd.HasSection(section) {
                etext := fmt.Sprintf("Duplicate section found: %s", section)
                returnedError = errors.New(etext)
                return blank, returnedError
            } else {
                cd.AddSection(section)
            }
        case cd.Regexp("confPair").MatchString(line):
            if section == "" {
                etext := fmt.Sprintf("Option pair not declared within a section: %s", line)
                return blank, errors.New(etext)
            }
            pair := cd.Regexp("confPair").FindStringSubmatch(line)
            key := pair[1]
            val := pair[2]
            if cd.HasSecOpt(section, key) {
                etext := fmt.Sprintf("Duplicate option %s found in section %s.", key, section)
                returnedError = errors.New(etext)
                return blank, returnedError
            } else {
                returnedError = cd.AddSecOpt(section, key, val)
                if returnedError != nil {
                    return blank, returnedError
                }
            }
        }
    }
    returnedError = cd.Interpolation()
    return cd, returnedError
}

func (c *ConfigData) Printer () {
    for section, secdata := range(c.Sections) {
        fmt.Println("Section:", section)
        for option, value := range(secdata.Options) {
            fmt.Printf("Option: %s == %s\n", option, value)
        }
    }
}
