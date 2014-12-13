package client

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	gnc "code.google.com/p/goncurses"
)

type Pot struct {
	c *DockerCli
}

// CommonLine contains information common to each printed line
type CommonLine struct {
	Id      string // PID or docker hash
	Command string // command name or CMD
	Uptime  string // container uptime or process uptime
	Status  string // container status or process state
	CPU     string // % of CPU used (if container, sum of % of processes)
	RAM     string // RAM used (if container, sum of RAM of processes)
}

// ContainerLine contains information specific to a docker container line
type ContainerLine struct {
	Name       string // docker container name
	Image      string // docker container image
	CommonLine        // same props as processes
}

func (c *ContainerLine) String() string {
	return fmt.Sprintf(strings.Join(
		[]string{c.Name, c.Image, c.Id, c.Command,
			c.Uptime, c.Status, c.CPU, c.RAM}, "\t"))
}

// ProcessLine contains information about a process
type ProcessLine ContainerLine

func (c *ProcessLine) String() string {
	return fmt.Sprintf(strings.Join(
		[]string{"", "", c.Id, c.Command,
			c.Uptime, c.Status, c.CPU, c.RAM}, "\t"))
}

type Container struct {
	container ContainerLine // information about the container
	processes []ProcessLine // information about the processes
}

var (
	containerTitle = "Name\tImage\tId\tCommand\tUptime\tStatus\tCPU\tRAM"
	active         = 0
	scroll         = 0
)

// Returns the list of running containers as well as internal processes
func (pot *Pot) Snapshot() {
	_, _, err := pot.c.call("GET", "/containers/", nil, false)
	if err != nil {
		return
	}
}

var cnts = []Container{
	{
		container: ContainerLine{
			Name:  "93%%",
			Image: "4242",
			CommonLine: CommonLine{
				Id: "12",
			},
		},
		processes: []ProcessLine{
			{
				Name:  "64%%",
				Image: "908",
				CommonLine: CommonLine{
					Id: "9",
				},
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
	{
		container: ContainerLine{
			Name:  "43.7%%",
			Image: "4385",
			CommonLine: CommonLine{
				Id: "1",
			},
		},
	},
}

func printActive(win *gnc.Window, s string, lc int, i int) {
	if i < scroll {
		return
	}
	if active == i {
		win.AttrOn(gnc.A_REVERSE)
		win.Println(s)
		win.AttrOff(gnc.A_REVERSE)
	} else {
		win.Println(s)
	}
}

func update(win *gnc.Window, lc int) {
	ss := make([]string, 0, 10)
	for _, cnt := range cnts {
		ss = append(ss, cnt.container.String())
		for _, proc := range cnt.processes {
			ss = append(ss, proc.String())
		}
	}
	if active < 0 {
		active++
	} else if active >= len(ss) {
		active--
	}
	if active > scroll+lc-1 {
		scroll++
	}
	if active < scroll {
		scroll--
	}
	for i, s := range ss {
		printActive(win, s, lc, i)
	}
}

func (pot *Pot) Run() {
	win, err := gnc.Init()
	win.Keypad(true)
	gnc.Echo(false)
	gnc.Cursor(0)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer gnc.End()

	k := make(chan gnc.Key)
	t := time.Tick(time.Second)

	go func(scr *gnc.Window, c chan gnc.Key) {
		for {
			c <- scr.GetChar()
		}
	}(win, k)
	for {
		my, _ := win.MaxYX()
		lc := my - 1 // size max of y - header (1)
		select {
		case kk := <-k:
			switch kk {
			case 'q':
				return
			case gnc.KEY_DOWN:
				active = active + 1
			case gnc.KEY_UP:
				active = active - 1
			}
		case <-t:
		}
		win.Erase()
		o, _ := exec.Command("uptime").Output()
		win.Printf("%s\n", o)

		win.AttrOn(gnc.A_REVERSE)
		win.Println(containerTitle)
		win.AttrOff(gnc.A_REVERSE)
		update(win, lc)
		win.Refresh()
	}
}

func NewPot(c *DockerCli) *Pot {
	return &Pot{c}
}
