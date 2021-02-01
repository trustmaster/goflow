package goflow

import (
	"fmt"
	"io/ioutil"
	"path"
	"plugin"
	"strings"
)

const plugInSuffix = `_goplug.so`

//PlugIn something
type PlugIn interface {
	Component
	SetParams(params map[string]interface{}) error
	GetParams() map[string]interface{}
	GetParam(param string) interface{}
}

//PlugInS something
type PlugInS struct {
	params  map[string]interface{}
	persist bool
}

func processForever(c Component) {
	for {
		c.Process()
	}
}

func process(c Component) {
	c.Process()
}

func (s *PlugInS) Process() {
	if s.persist {
		processForever(s)
	} else {
		process(s)
	}
}

func (s *PlugInS) SetParams(params map[string]interface{}) error {
	s.params = params
	value, ok := params["persist"].(bool)
	if ok {
		s.persist = value
	} else {
		s.persist = false
	}
	return nil
}

func (s *PlugInS) GetParams() map[string]interface{} {
	return s.params
}

func (s *PlugInS) GetParam(param string) interface{} {
	return s.params[param]
}

// LoadComponents goes through all paths, opens all plugins in those paths
// and loads them into factory
// Plugins are denoted by *_goplug.so, The filename must begin with a capitolized letter
func LoadComponents(paths []string, factory *Factory) ([]string, error) {
	var loaded []string

	for _, apath := range paths {
		//fmt.Println("Loading plugins at", apath)
		files, err := ioutil.ReadDir(apath)
		if err != nil {
			fmt.Printf("Path %s not found, error=%s.", apath, err.Error())
			continue
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), plugInSuffix) {
				plugpath := path.Join(apath, file.Name())
				plug, err := plugin.Open(plugpath)
				if err != nil {
					fmt.Printf("Can't open plugin %s, error=%s.", plugpath, err.Error())
					continue
				}
				//get name from name - _goplug.so and register with contructor
				name := strings.TrimSuffix(file.Name(), plugInSuffix)
				symbol, err := plug.Lookup(name)
				if err != nil {
					fmt.Printf("Can't find symbol %s in plugin %s, error=%s.", name, plugpath, err.Error())
					continue
				}
				constructor := symbol.(func() (interface{}, error))
				anerr := factory.Register(name, constructor)
				if anerr != nil {
					fmt.Printf("Failed to register plugin %s, error=%s.", plugpath, err.Error())
					continue
				}
				loaded = append(loaded, name)
			}
		}
	}
	return loaded, nil
}
