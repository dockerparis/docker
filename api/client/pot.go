package client

import (
	"fmt"
	"time"
)

type Pot struct {
	c *DockerCli
}

// Returns the list of running containers as well as internal processes
func (pot *Pot) Snapshot() {
	_, _, err := pot.c.call("GET", "/containers/", nil, false)
	if err != nil {
		return
	}
	fmt.Printf("salut mec\n")
}

func (pot *Pot) Run() {
	for {
		pot.Snapshot()
		time.Sleep(1 * 1e9)
	}
}

func NewPot(c *DockerCli) *Pot {
	return &Pot{c}
}
