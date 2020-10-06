package main

import (
	"github.com/dahvid/goflow"
)

type adder struct {
	Left  <-chan int
	Right <-chan int
	Out   chan<- int
	goflow.PlugInS
}

func (c *adder) Process() {
	for {
		x := <-c.Left
		y := <-c.Right
		r := x + y
		//fmt.Println("result=", r)
		c.Out <- r
	}

}

//Adder somthin
func Adder() (interface{}, error) {
	//fmt.Println("Creating new adder")
	return new(adder), nil
}
