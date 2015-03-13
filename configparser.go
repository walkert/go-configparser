package configparser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type InterpolateObj struct {
	value      string
	refSection string
}

type SectionData struct {
	options     map[string]string
	interpolate map[string]InterpolateObj
}

// Option interface for declaring expected config options
type Option interface {
	Set(s string) error
	Required() bool
}

type IntOpt struct {
	Name    string
	Req     bool
	Section string
	Value   *int
}

func (i IntOpt) Set(v string) error {
	val, err := strconv.ParseInt(v, 0, 64)
	*i.Value = int(val)
	return err
}

func (i IntOpt) Required() bool {
	return i.Req
}

type StringOpt struct {
	Name    string
	Req     bool
	Section string
	Value   *string
}

func (s StringOpt) Set(v string) error {
	*s.Value = v
	return nil
}

func (s StringOpt) Required() bool {
	return s.Req
}

type BoolOpt struct {
	Name    string
	Req     bool
	Section string
	Value   *bool
}

func (b BoolOpt) Set(s string) error {
	val, err := strconv.ParseBool(s)
	if err != nil {
		etext := fmt.Sprintf("Invalid value for boolean option: %s", s)
		return errors.New(etext)
	}
	*b.Value = val
	return nil
}

func (b BoolOpt) Required() bool {
	return b.Req
}

type ListOpt struct {
	Name    string
	Req     bool
	Section string
	Value   *[]string
}

func (l ListOpt) Set(v string) error {
	listSep := regexp.MustCompile(`\s*,\s*`)
	list := listSep.Split(v, -1)
	if len(list) == 0 {
		etext := fmt.Sprintf("Invalid list value: %s", v)
		return errors.New(etext)
	}
	*l.Value = list
	return nil
}

func (l ListOpt) Required() bool {
	return l.Req
}

// addOption adds a new option to a given section
func addOption(c *ConfigParser, section, option, value string) {
	c.sections[section].options[option] = value
}

// addInter creats an interpolation object. Which can be used to
// map interpolation strings such as %global(path) to an InterpolateObj
func addInter(c *ConfigParser, section, option string, idata []string) (retval error) {
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

// The ConfigParser struct represents the configuration to be parsed
// fname is the name of the config file to be processed
// parsed sets whether the file has been parsed or not
// regexps is a map of regexps used for config parsing
type ConfigParser struct {
	fname    string
	parsed   bool
	regexps  map[string]*regexp.Regexp
	sections map[string]SectionData
	Options  map[string]Option
}

func NewConfigParser(fname string) (cd ConfigParser) {
	secdata := make(map[string]SectionData)
	regexps := make(map[string]*regexp.Regexp)
	options := make(map[string]Option, 0)
	regexps["inter"] = regexp.MustCompile(`%(\w*)\((.*)\)`)
	regexps["secMatch"] = regexp.MustCompile(`^\[(.*)\]$`)
	regexps["confPair"] = regexp.MustCompile(`^(\w*)\s+=\s+(.*)`)
	return ConfigParser{fname, false, regexps, secdata, options}
}

// HasSection reports whether the named section exists
func (c *ConfigParser) HasSection(section string) bool {
	if _, ok := c.sections[section]; ok {
		return true
	} else {
		return false
	}
}

// HasSection reports whether the named option exists in the section
func (c *ConfigParser) HasOption(section, option string) bool {
	if c.HasSection(section) {
		if _, ok := c.sections[section].options[option]; ok {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

// regexp returns the Regexp pointer for the given name
func (c *ConfigParser) regexp(regexp string) *regexp.Regexp {
	return c.regexps[regexp]
}

// GetOption returns the named option from the section
func (c *ConfigParser) GetOption(section, option string) (value string, retval error) {
	retval = nil
	value = ""
	if c.HasSection(section) {
		if c.HasOption(section, option) {
			value = c.sections[section].options[option]
		} else {
			etext := fmt.Sprintf("Option '%s' does not exist in section '%s'.\n", option, section)
			retval = errors.New(etext)
		}
	} else {
		etext := fmt.Sprintf("Section '%s' does not exist.\n", section)
		retval = errors.New(etext)
	}
	return
}

// setSecOpt sets the named option for the section
func (c *ConfigParser) setSecOpt(section, option, value string) {
	c.sections[section].options[option] = value
}

// addSecOpt adds the option, value pair to the named section
// If an interpolation string is found addInter is called to add
// the interpolation value.
func (c *ConfigParser) addSecOpt(section, option, value string) (retval error) {
	retval = nil
	addOption(c, section, option, value)
	igroup := c.regexp("inter").FindStringSubmatch(value)
	if len(igroup) == 3 {
		retval = addInter(c, section, option, igroup[1:])
	}
	return retval
}

// addSecOpt adds the named section to the ConfigParser object
func (c *ConfigParser) addSection(section string) {
	options := make(map[string]string)
	interpolate := make(map[string]InterpolateObj)
	sd := SectionData{options, interpolate}
	c.sections[section] = sd
}

// GetSections returns a slice of section names
func (c *ConfigParser) GetSections() (sectionList []string) {
	sectionList = make([]string, len(c.sections))
	for key, _ := range c.sections {
		sectionList = append(sectionList, key)
	}
	return
}

// interpolation iterates over every section and attempts to interpolate
// any options which contain an interpolation marker
func (c *ConfigParser) interpolation() error {
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
						etext := fmt.Sprintf("Cannot interpolate '%s', '%s' section does not contain key '%s'.", key, refsection, irefval.value)
						return errors.New(etext)
					}
				} else {
					etext := fmt.Sprintf("Cannot interpolate '%s', '%s' section not defined.", key, refsection)
					return errors.New(etext)
				}
			}
		}
	}
	return retval
}

// interpolate resolves an interpolation marker into a valid option
// For example this converts %global(mpath)/logs -> /master/path/logs
func (c *ConfigParser) interpolate(section, key, refsection, irefkey string) {
	current, _ := c.GetOption(section, key)
	ival, _ := c.GetOption(refsection, irefkey)
	replacement := c.regexp("inter").ReplaceAllLiteralString(current, ival)
	c.setSecOpt(section, key, replacement)
}

// Parse does main processing of the configuration file. It discovers all
// sections and options and creates the relevant objects.
// When that stage it complete, it then iterates of the Options slice and
// if any options are found attempts to set their values.
func (c *ConfigParser) Parse() (err error) {
	err = nil
	fd, err := os.Open(c.fname)
	if err != nil {
		return
	}
	defer fd.Close()
	var section string
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case c.regexp("secMatch").MatchString(line):
			section = c.regexp("secMatch").FindStringSubmatch(line)[1]
			if c.HasSection(section) {
				etext := fmt.Sprintf("Duplicate section found: '%s'", section)
				err = errors.New(etext)
				return
			} else {
				c.addSection(section)
			}
		case c.regexp("confPair").MatchString(line):
			if section == "" {
				etext := fmt.Sprintf("Option pair not declared within a section: '%s'", line)
				err = errors.New(etext)
				return
			}
			pair := c.regexp("confPair").FindStringSubmatch(line)
			key := pair[1]
			val := pair[2]
			if c.HasOption(section, key) {
				etext := fmt.Sprintf("Duplicate option '%s' found in section '%s'.", key, section)
				err = errors.New(etext)
				return
			} else {
				err = c.addSecOpt(section, key, val)
				if err != nil {
					return
				}
			}
		}
	}
	err = c.interpolation()
	if err != nil {
		return
	}
	c.parsed = true
	flat, err := c.GetFlatConfigMap()
	if err != nil {
		return
	}
	// Iterate over the list of registered options, see if they're set
	var etext, value string
	for key, option := range c.Options {
		klist := strings.Split(key, "-")
		if len(klist) == 1 {
			// If this is a top-level option, check for it in flat. Error
			// only if it's Required value is True
			if value, ok := flat[key]; !ok {
				if option.Required() {
					etext = fmt.Sprintf("Top-level option %s has not been set but is required.", key)
					err = errors.New(etext)
					return
				}
			} else {
				option.Set(value)
			}
		} else {
			name := klist[0]
			section := klist[1]
			if !c.HasSection(section) {
				fmt.Println("No section: ", section)
				if option.Required() {
					etext = fmt.Sprintf("Option %s was expected in section %s but this section does not exist.", name, section)
					err = errors.New(etext)
					return
				}
			} else {
				value, err = c.GetOption(section, name)
				if err != nil {
					if option.Required() {
						etext = fmt.Sprintf("Option %s does not exist in section %s.", name, section)
						err = errors.New(etext)
						return
					}
				} else {
					option.Set(value)
				}
			}
		}
	}
	return
}

// GetConfigMap returns a map[string]map[string]string which contains the option/value
// pairs for each config section.
//
func (c *ConfigParser) GetConfigMap() (confmap map[string]map[string]string, err error) {
	confmap = make(map[string]map[string]string)
	err = c.Parse()
	if err != nil {
		return
	}
	for section, secdata := range c.sections {
		confmap[section] = make(map[string]string)
		for option, value := range secdata.options {
			confmap[section][option] = value
		}
	}
	return
}

// GetFlatConfigMap collapses the config into a simple map of option/value pairs.
// A check must be performed for duplicate options again since this is possible
// across sections.
func (c *ConfigParser) GetFlatConfigMap() (flatmap map[string]string, err error) {
	flatmap = make(map[string]string)
	if !c.parsed {
		err = c.Parse()
		if err != nil {
			return
		}
	}
	for _, secdata := range c.sections {
		for option, value := range secdata.options {
			if _, ok := flatmap[option]; ok {
				etext := fmt.Sprintf("Cannot create flat config map, duplicate option found: %s.", option)
				err = errors.New(etext)
				return
			}
			flatmap[option] = value
		}
	}
	return
}

// buildKey returns an option key. If the option specifies a section
// the key is in the form section-name, otherwise, just name.
func buildKey(name, section string) (key string) {
	if section == "" {
		key = name
	} else {
		key = fmt.Sprintf("%s-%s", name, section)
	}
	return
}

func (c *ConfigParser) IntOption(name, section string, def int, required bool) *int {
	p := new(int)
	*p = def
	key := buildKey(name, section)
	c.Options[key] = IntOpt{Name: name, Req: required, Section: section, Value: p}
	return p
}

func (c *ConfigParser) BoolOption(name, section string, def, required bool) *bool {
	p := new(bool)
	*p = def
	key := buildKey(name, section)
	c.Options[key] = BoolOpt{Name: name, Req: required, Section: section, Value: p}
	return p
}

func (c *ConfigParser) StringOption(name, section, def string, required bool) *string {
	p := new(string)
	*p = def
	key := buildKey(name, section)
	c.Options[key] = StringOpt{Name: name, Req: required, Section: section, Value: p}
	return p
}

func (c *ConfigParser) ListOption(name, section string, required bool) *[]string {
	p := make([]string, 0)
	key := buildKey(name, section)
	c.Options[key] = ListOpt{Name: name, Req: required, Section: section, Value: &p}
	return &p
}
