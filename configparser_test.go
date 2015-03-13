package configparser

import (
	"testing"
)

func TestGetConfigMap(t *testing.T) {
	/* Return a map in the form map[string]map[string]string which
	   contain a map of each option->value pair for each section.
	*/
	cp := NewConfigParser("examples/basic.cfg")
	confMap, err := cp.GetConfigMap()
	if err != nil {
		t.Fatal("Error getting config map: ", err)
	}
	if len(confMap) != 2 {
		t.Fatal("Section count incorrect for standard config map")
	}
}

func TestGetFlatMap(t *testing.T) {
	/* Return a map in the form map[string]map[string]string which
	   contain a map of each option->value pair for each section.
	*/
	// fmt.Println("Create a flat config map object..")
	cp := NewConfigParser("examples/interpolate.cfg")
	confMap, err := cp.GetFlatConfigMap()
	t.Logf("inter: %+v\n", cp)
	if err != nil {
		t.Fatal("Error getting flat config map")
	}
	if len(confMap) != 6 {
		t.Fatalf("Config count incorrect for flat config map. Should be 6 but is %d.", len(confMap))
	}
}

func TestHasSection(t *testing.T) {
	// Test for the presence of a known section
	cp := NewConfigParser("examples/interpolate.cfg")
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	if !cp.HasSection("app1") {
		t.Fatal("Can't find 'app1' section")
	}
}

func TestHasOption(t *testing.T) {
	// Test for the presence of a known section/option
	cp := NewConfigParser("examples/basic.cfg")
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	if !cp.HasOption("app1", "port") {
		t.Fatal("Can't find 'port' option in the 'app1' section")
	}
}

func TestGetOption(t *testing.T) {
	// Try and get a non-existent option from a valid section
	cp := NewConfigParser("examples/basic.cfg")
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	_, err = cp.GetOption("app1", "badOpt")
	if err == nil {
		t.Fatal("Expected error from GetOption(), got 'nil'")
	}
	if err.Error() != "Option 'badOpt' does not exist in section 'app1'.\n" {
		t.Fatal("Unexpected error while getting 'bad' option from config")
	}
}

func TestRegsiterBool(t *testing.T) {
	cp := NewConfigParser("examples/basic.cfg")
	mybool := cp.BoolOption("running", "app1", false, true)
	if *mybool {
		t.Fatal("mybool should be false (default) prior to running Parse() but was true.")
	}
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	if !*mybool {
		t.Fatal("Var 'mybool' should be true!")
	}
}

func TestListOpt(t *testing.T) {
	cp := NewConfigParser("examples/basic.cfg")
	mylist := cp.ListOption("apps", "global", false)
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	if len(*mylist) != 2 {
		t.Fatalf("Expected 2 list items but got: %d", len(*mylist))
	}
	if (*mylist)[0] != "app1" {
		t.Fatalf("Expected item 0 to be 'app1' but got: %s", (*mylist)[0])
	}
}

func TestStringOpt(t *testing.T) {
	cp := NewConfigParser("examples/basic.cfg")
	mystring := cp.StringOption("master", "global", "def_string", false)
	if *mystring != "def_string" {
		t.Fatal("Var 'mystring' should be 'def_string' (default) prior to running Parse() but was: %s", *mystring)
	}
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	if *mystring != "/opt/applications" {
		t.Fatalf("Expected var 'mystring' to be '/opt/applications' but got '%s'.", *mystring)
	}
}

func TestIntOpt(t *testing.T) {
	cp := NewConfigParser("examples/basic.cfg")
	myint := cp.IntOption("port", "app1", 1000, false)
	if *myint != 1000 {
		t.Fatal("Var 'myint' should be '1000' (default) prior to running Parse() but was: %d", *myint)
	}
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	if *myint != 8000 {
		t.Fatalf("Expected var 'mystring' to be '8000' but got '%d'.", *myint)
	}
}

func TestNotSettingRequiredOption(t *testing.T) {
	cp := NewConfigParser("examples/basic.cfg")
	_ = cp.StringOption("not_there", "global", "", true)
	err := cp.Parse()
	if err == nil {
		t.Fatal("Expecteded an error from Parse() but none.")
	} else {
		if err.Error() != "Option not_there does not exist in section global." {
			t.Fatal("Unexpected error output: ", err)
		}
	}
}

func TestFlatStringOpt(t *testing.T) {
	cp := NewConfigParser("examples/basic.cfg")
	mystring := cp.StringOption("master", "", "", false)
	err := cp.Parse()
	if err != nil {
		t.Fatalf("Got unexpected error from Parse(): %s", err)
	}
	if *mystring != "/opt/applications" {
		t.Fatalf("Var 'mybool' expected to be '/opt/applications' but got '%s'.", *mystring)
	}
}
