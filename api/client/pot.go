package client

import (
	"fmt"
	"time"
)

type Pot struct {
	c *DockerCli
}

// CommonLine contains information common to each printed line
type CommonLine struct {
	Id string		// PID or docker hash
	Command string		// command name or CMD
	Uptime string		// container uptime or process uptime
	Status string		// container status or process state
	CPU string		// % of CPU used (if container, sum of % of processes)
	RAM string		// RAM used (if container, sum of RAM of processes)
}

// ContainerLine contains information specific to a docker container line
type ContainerLine struct {
	Name string		// docker container name
	Image string		// docker container image
	CommonLine		// same props as processes
}

// ProcessLine contains information about a process
type ProcessLine ContainerLine

type Container struct {
	container ContainerLine	// information about the container	
	processes []ProcessLine	// information about the processes
}

// Returns the list of running containers as well as internal processes
func (pot *Pot) Snapshot() {
	_, _, err := pot.c.call("GET", "/containers/", nil, false)
	if err != nil {
		return
	}
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
