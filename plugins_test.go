package goflow

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestBadParams(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	paths := []string{"boogie"}
	_, err := LoadComponents(paths, NewFactory())
	if err != nil {
		t.Error("LoadComponenets succeeded with bad parameters")
	}
	paths = []string{"/usr/lib", "/lib"}
	_, err = LoadComponents(paths, NewFactory())
	if err != nil {
		t.Error("LoadComponenets succeeded with bad parameters")
	}
}

func createFactory(t *testing.T) *Factory {
	testPlugs := path.Join(os.Getenv("GOPATH"), "bin")
	paths := []string{testPlugs}
	factory := NewFactory()
	plugs, err := LoadComponents(paths, factory)
	if err != nil {
		t.Error("Failed loading compononents in ./test_plugins", err)
	}
	if len(plugs) == 0 {
		t.Error("No plugins found in directory", testPlugs)
	}
	//fmt.Println("Opened plugs", plugs)
	return factory
}

func TestOpening(t *testing.T) {
	factory := createFactory(t)
	any, err := factory.Create("Plug1")
	if err != nil {
		t.Error("Failed to create object Plug1", err)
	}
	plug1 := any.(PlugIn)
	recieved := plug1.Info()
	expected := map[string]string{"One": "Two"}

	if !reflect.DeepEqual(recieved, expected) {
		t.Error("Recieved bad meta data expected", expected, "got", recieved)
	}
	//fmt.Println(plug1.Info())
}

var testJSON = `{
	"properties": {
		"name": "testJSON"
	},
	"processes": {
		"gen1": {
			"component": "NGen"
		},
		"gen2": {
			"component": "NGen"
		},
		"adder": {
			"component": "Adder"
		}
	},
	"connections": [
		{
			"tgt": {
				"process": "gen1",
				"port": "Init"
			}
		},
		{
			"tgt": {
				"process": "gen2",
				"port": "Init"
			}
		},
		{
			"src": {
				"process": "gen1",
				"port": "Out"
			},
			"tgt": {
				"process": "adder",
				"port": "Left"
			}
		},
		{
			"src": {
				"process": "gen2",
				"port": "Out"
			},
			"tgt": {
				"process": "adder",
				"port": "Right"
			}
		}
	],
	"exports": [
		{
			"private": "adder.Out",
			"public": "Out"
		},
		{
			"private": "gen1.Init",
			"public" : "In1"
		},
		{
			"private": "gen2.Init",
			"public" : "In2"
		}
	]
}`

//TestGraph something
func TestLoadGraph(t *testing.T) {
	factory := createFactory(t)
	net := ParseJSON([]byte(testJSON), factory)
	if net == nil {
		t.Error("Could not load JSON")
	}

	in1 := make(chan [10]int)
	in2 := make(chan [10]int)
	out := make(chan int)

	net.SetInPort("In1", in1)
	net.SetInPort("In2", in2)
	net.SetOutPort("Out", out)

	Run(net)

	in1 <- [10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	in2 <- [10]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	// Wait for the network setup
	//<-net.Ready()

	// Close start to halt it normally
	//close(start)

	test := [10]int{11, 22, 33, 44, 55, 66, 77, 88, 99, 110}

	for _, v := range test {
		result := <-out
		if result != v {
			t.Errorf("Wrong results: expected %d got %d", v, result)
		}
	}
	//<-wait
}
