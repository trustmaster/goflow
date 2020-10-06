package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dahvid/goflow"
)

// RandomNumberGenerator Listen for ssh connection
type RandomNumberGenerator struct {
	Generator *rand.Rand
	Out       chan<- int //output
	goflow.PlugInS
}

// Process listen
func (c *RandomNumberGenerator) Process() {
	for {
		i := c.Generator.Intn(100)
		fmt.Println("Generating", i)
		c.Out <- i
	}
}

//Plug1 something
func Plug1() (interface{}, error) {
	seed := rand.NewSource(time.Now().UnixNano())
	gen := rand.New(seed)
	r := new(RandomNumberGenerator)
	r.Generator = gen
	return r, nil
}
