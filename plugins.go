package goflow

import (
	"io/ioutil"
	"log"
	"path"
	"plugin"
	"strings"
)

const plugInSuffix = `_goplug.so`

//PlugIn something
type PlugIn interface {
	Component
	Info() map[string]string
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
			log.Printf("Path %s not found, error=%s.", apath, err.Error())
			continue
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), plugInSuffix) {
				plugpath := path.Join(apath, file.Name())
				plug, err := plugin.Open(plugpath)
				if err != nil {
					log.Printf("Can't open plugin %s, error=%s.", plugpath, err.Error())
					continue
				}
				//get name from name - _goplug.so and register with contructor
				name := strings.TrimSuffix(file.Name(), plugInSuffix)
				symbol, err := plug.Lookup(name)
				if err != nil {
					log.Printf("Can't find symbol %s in plugin %s, error=%s.", name, plugpath, err.Error())
					continue
				}
				constructor := symbol.(func() (interface{}, error))
				anerr := factory.Register(name, constructor)
				if anerr != nil {
					log.Printf("Failed to register plugin %s, error=%s.", plugpath, err.Error())
					continue
				}
				loaded = append(loaded, name)
			}
		}
	}
	return loaded, nil
}
