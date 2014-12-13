package client

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	gnc "code.google.com/p/goncurses"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/pkg/units"
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
func (pot *Pot) Snapshot() []Container {
	res := make([]Container, 0, 16)

	v := url.Values{}
	v.Set("all", "1")
	body, _, err := readBody(pot.c.call("GET", "/containers/json?"+v.Encode(), nil, false))
	if err != nil {
		fmt.Printf("readBody failed %v\n", err)
		return res
	}
	outs := engine.NewTable("Created", 0)
	if _, err = outs.ReadListFrom(body); err != nil {
		fmt.Printf("what?\n")
		return res
	}
	for _, out := range outs.Data {
		var c Container

		c.container.Id = out.Get("Id")
		c.container.Command = strconv.Quote(out.Get("Command"))
		c.container.Image = "Soon"
		c.container.Name = out.GetList("Names")[0]
		c.container.Uptime = units.HumanDuration(time.Now().UTC().Sub(time.Unix(out.GetInt64("Created"), 0)))
		c.container.Status = out.Get("Status")

		res = append(res, c)
	}

	return res
}

func printActive(win *gnc.Window, s string, lc int, i int) {
	if i < scroll || i >= scroll+lc {
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

func (pot *Pot) update(win *gnc.Window, lc int) {
	ss := make([]string, 0, 10)
	for _, cnt := range pot.Snapshot() {
		ss = append(ss, cnt.container.String())
		for _, proc := range cnt.processes {
			ss = append(ss, proc.String())
		}
	}
	if active < 0 {
		active = 0
	} else if active >= len(ss) {
		active = len(ss) - 1
	}
	if active >= scroll+lc {
		scroll = active - lc + 1
	}
	if active < scroll {
		scroll = active
	}
	for i, s := range ss {
		printActive(win, s, lc, i)
	}
}

func (pot *Pot) Run() {
	win, err := gnc.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer gnc.End()
	win.Keypad(true)
	gnc.Echo(false)
	gnc.Cursor(0)

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGWINCH)

	k := make(chan gnc.Key)
	t := time.Tick(time.Second)

	go func(scr *gnc.Window, c chan gnc.Key) {
		for {
			c <- scr.GetChar()
		}
	}(win, k)

	for {
		my, _ := win.MaxYX()
		lc := my - 2 // size max of y - header (1)
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
		case <-s:
			gnc.End()
			win.Refresh()
		}
		win.Erase()
		o, _ := exec.Command("uptime").Output()
		win.Printf("%s", o)

		win.AttrOn(gnc.A_REVERSE)
		win.Println(containerTitle)
		win.AttrOff(gnc.A_REVERSE)
		pot.update(win, lc)
		win.Refresh()
	}
}

func NewPot(c *DockerCli) *Pot {
	return &Pot{c}
}
