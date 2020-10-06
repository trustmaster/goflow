package main

import (
	"math"

	"github.com/dahvid/goflow"
)

//randomNumberGenerator something
type generator struct {
	Out chan<- int //output
	goflow.PlugInS
}

// Process listen
func (c *generator) Process() {

	values := c.GetParam("inputs")
	//fmt.Println("got values", values)
	//fmt.Println("go type", reflect.TypeOf(values))
	for _, v := range values.([]interface{}) {
		//fmt.Println("Generating", v.(float64))
		c.Out <- int(math.Round(v.(float64)))
	}
}

//NGen something
func NGen() (interface{}, error) {
	return new(generator), nil
}
