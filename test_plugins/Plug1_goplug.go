package main

import (
	"fmt"
	"math/rand"
	"time"
)

// RandomNumberGenerator Listen for ssh connection
type RandomNumberGenerator struct {
	Generator *rand.Rand
	meta      map[string]string
	Out       chan<- int //output
}

// Process listen
func (c *RandomNumberGenerator) Process() {
	for {
		i := c.Generator.Intn(100)
		fmt.Println("Generating", i)
		c.Out <- i
	}
}

//Info something
func (c *RandomNumberGenerator) Info() map[string]string {
	return c.meta
}

//Plug1 something
func Plug1() (interface{}, error) {
	seed := rand.NewSource(time.Now().UnixNano())
	gen := rand.New(seed)
	r := new(RandomNumberGenerator)
	r.Generator = gen
	r.meta = make(map[string]string)
	r.meta["One"] = "Two"
	return r, nil
}
