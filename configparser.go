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
    RefSection string
}

type SectionData struct {
    Options map[string]string
    Interpolate map[string]InterpolateObj
}

func addOption (c *ConfigData, section, option, value string) {
    c.Sections[section].Options[option] = value
}

func addInter (c *ConfigData, section, option string, idata []string) error {
    var retval error = nil
    refsection := idata[0]
    value := idata[1]
    if refsection == "local" {
        refsection = section
    }
    inter := InterpolateObj{value, refsection}
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
    regexps["inter"] = regexp.MustCompile(`%(\w*)\((.*)\)`)
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
func (c *ConfigData) HasOption (section, option string) bool {
    if _, ok := c.Sections[section].Options[option] ; ok {
        return true
    } else {
        return false
    }
}

func (c *ConfigData) regexp (regexp string) *regexp.Regexp {
    return c.Regexps[regexp]
}

func (c *ConfigData) GetOption (section, option string) string {
    return c.Sections[section].Options[option]
}

func (c *ConfigData) setSecOpt (section, option, value string) {
    c.Sections[section].Options[option] = value
}

func (c *ConfigData) addSecOpt (section, key, value string) error {
    var retval error = nil
    addOption(c, section, key, value)
    igroup := c.regexp("inter").FindStringSubmatch(value)
    if len(igroup) == 3 {
        retval = addInter(c, section, key, igroup[1:])
    }
    return retval
}

func (c *ConfigData) addSection (section string) {
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

func (c *ConfigData) interpolation () error {
    var retval error = nil
    var refsection string
    for _, section := range(c.GetSections()) {
        if len(c.Sections[section].Interpolate) > 0 {
            for key, irefval := range(c.Sections[section].Interpolate) {
                refsection = irefval.RefSection
                if c.HasSection(refsection) {
                    if c.HasOption(refsection, irefval.Value) {
                        c.interpolate(section, key, refsection, irefval.Value)
                    } else { 
                        etext := fmt.Sprintf("Cannot interpolate %s, %s section does not contain key %s.", key, refsection, irefval.Value)
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

func (c *ConfigData) interpolate (section, key, refsection, irefkey string) {
    current := c.GetOption(section, key)
    ival := c.GetOption(refsection, irefkey)
    replacement := c.regexp("inter").ReplaceAllLiteralString(current, ival)
    c.setSecOpt(section, key, replacement)
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
        case cd.regexp("secMatch").MatchString(line):
            section = cd.regexp("secMatch").FindStringSubmatch(line)[1]
            if cd.HasSection(section) {
                etext := fmt.Sprintf("Duplicate section found: %s", section)
                returnedError = errors.New(etext)
                return blank, returnedError
            } else {
                cd.addSection(section)
            }
        case cd.regexp("confPair").MatchString(line):
            if section == "" {
                etext := fmt.Sprintf("Option pair not declared within a section: %s", line)
                return blank, errors.New(etext)
            }
            pair := cd.regexp("confPair").FindStringSubmatch(line)
            key := pair[1]
            val := pair[2]
            if cd.HasOption(section, key) {
                etext := fmt.Sprintf("Duplicate option %s found in section %s.", key, section)
                returnedError = errors.New(etext)
                return blank, returnedError
            } else {
                returnedError = cd.addSecOpt(section, key, val)
                if returnedError != nil {
                    return blank, returnedError
                }
            }
        }
    }
    returnedError = cd.interpolation()
    if returnedError != nil {
        cd = blank
    }
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
