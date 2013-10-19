package configparser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
)

type InterpolateObj struct {
	value      string
	refSection string
}

type SectionData struct {
	options     map[string]string
	interpolate map[string]InterpolateObj
}

func addOption(c *ConfigData, section, option, value string) {
	c.sections[section].options[option] = value
}

func addInter(c *ConfigData, section, option string, idata []string) (retval error) {
	retval = nil
	refsection := idata[0]
	value := idata[1]
	if refsection == "local" {
		refsection = section
	}
	inter := InterpolateObj{value, refsection}
	c.sections[section].interpolate[option] = inter
	return
}

type ConfigData struct {
	sections map[string]SectionData
	regexps  map[string]*regexp.Regexp
}

func NewConfigData() ConfigData {
	secdata := make(map[string]SectionData)
	regexps := make(map[string]*regexp.Regexp)
	regexps["inter"] = regexp.MustCompile(`%(\w*)\((.*)\)`)
	regexps["secMatch"] = regexp.MustCompile(`^\[(.*)\]$`)
	regexps["confPair"] = regexp.MustCompile(`^(\w*)\s+=\s+(.*)`)
	return ConfigData{secdata, regexps}
}

func (c *ConfigData) HasSection(section string) bool {
	if _, ok := c.sections[section]; ok {
		return true
	} else {
		return false
	}
}
func (c *ConfigData) HasOption(section, option string) bool {
	if _, ok := c.sections[section].options[option]; ok {
		return true
	} else {
		return false
	}
}

func (c *ConfigData) regexp(regexp string) *regexp.Regexp {
	return c.regexps[regexp]
}

func (c *ConfigData) GetOption(section, option string) (value string, retval error) {
	retval = nil
	value = ""
	if c.HasOption(section, option) {
		value = c.sections[section].options[option]
	} else {
		etext := fmt.Sprintf("Option %s does not exist in section %s.\n", option, section)
		retval = errors.New(etext)
	}
	return
}

func (c *ConfigData) setSecOpt(section, option, value string) {
	c.sections[section].options[option] = value
}

func (c *ConfigData) addSecOpt(section, key, value string) (retval error) {
	retval = nil
	addOption(c, section, key, value)
	igroup := c.regexp("inter").FindStringSubmatch(value)
	if len(igroup) == 3 {
		retval = addInter(c, section, key, igroup[1:])
	}
	return retval
}

func (c *ConfigData) addSection(section string) {
	options := make(map[string]string)
	interpolate := make(map[string]InterpolateObj)
	sd := SectionData{options, interpolate}
	c.sections[section] = sd
}

func (c *ConfigData) GetSections() (sectionList []string) {
	sectionList = make([]string, len(c.sections))
	for key, _ := range c.sections {
		sectionList = append(sectionList, key)
	}
	return
}

func (c *ConfigData) interpolation() error {
	var retval error = nil
	var refsection string
	for _, section := range c.GetSections() {
		if len(c.sections[section].interpolate) > 0 {
			for key, irefval := range c.sections[section].interpolate {
				refsection = irefval.refSection
				if c.HasSection(refsection) {
					if c.HasOption(refsection, irefval.value) {
						c.interpolate(section, key, refsection, irefval.value)
					} else {
						etext := fmt.Sprintf("Cannot interpolate %s, %s section does not contain key %s.", key, refsection, irefval.value)
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

func (c *ConfigData) interpolate(section, key, refsection, irefkey string) {
	current, _ := c.GetOption(section, key)
	ival, _ := c.GetOption(refsection, irefkey)
	replacement := c.regexp("inter").ReplaceAllLiteralString(current, ival)
	c.setSecOpt(section, key, replacement)
}

func Parse(fname string) (ConfigData, error) {
	var retval error = nil
	blank := NewConfigData()
	cd := NewConfigData()
	fd, err := os.Open(fname)
	if err != nil {
		retval = err
		return cd, retval
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
				retval = errors.New(etext)
				return blank, retval
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
				retval = errors.New(etext)
				return blank, retval
			} else {
				retval = cd.addSecOpt(section, key, val)
				if retval != nil {
					return blank, retval
				}
			}
		}
	}
	retval = cd.interpolation()
	if retval != nil {
		cd = blank
	}
	return cd, retval
}

func (c *ConfigData) GetConfigMap() (confmap map[string]map[string]string) {
	confmap = make(map[string]map[string]string)
	for section, secdata := range c.sections {
		confmap[section] = make(map[string]string)
		for option, value := range secdata.options {
			confmap[section][option] = value
		}
	}
	return
}

func (c *ConfigData) GetFlatConfigMap() (flatmap map[string]string, retval error) {
	flatmap = make(map[string]string)
	for _, secdata := range c.sections {
		for option, value := range secdata.options {
			if _, ok := flatmap[option]; ok {
				etext := fmt.Sprintf("Cannot create flat config map, duplicate option found: %s.", option)
				retval = errors.New(etext)
				return
			}
			flatmap[option] = value
		}
	}
	return
}
