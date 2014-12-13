package client

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"
	"time"

	"code.google.com/p/goncurses"
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

// ProcessLine contains information about a process
type ProcessLine ContainerLine

type Containe struct {
	container ContainerLine // information about the container
	processes []ProcessLine // information about the processes
}

// Returns the list of running containers as well as internal processes
func (pot *Pot) Snapshot() {
	_, _, err := pot.c.call("GET", "/containers/", nil, false)
	if err != nil {
		return
	}
}

type Container struct {
	Cpu   string
	Port  string
	State string
}

var m = map[string]Container{
	"pomme":   Container{"93%", "4242", "running"},
	"poire":   Container{"43.7%", "9754", "killed"},
	"haricot": Container{"72.5%", "3452", "sleeping"},
}

func test(stdscr *goncurses.Window, c chan goncurses.Key) {
	for {
		c <- stdscr.GetChar()
	}
}

func key(k goncurses.Key) bool {
	switch k {
	case 'q':
		return false
	default:
		return true
	}

}

func (pot *Pot) Run() {
	stdscr, err := goncurses.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer goncurses.End()
	k := make(chan goncurses.Key)
	t := time.Tick(time.Second)

	go test(stdscr, k)
	for {
		stdscr.Erase()
		select {
		case kk := <-k:
			stdscr.Println("key pressed!")
			if !key(kk) {
				return
			}
		case <-t:
			o, _ := exec.Command("uptime").Output()
			stdscr.Printf("%s\n", o)

			b := new(bytes.Buffer)
			t := new(tabwriter.Writer)
			t.Init(b, 0, 8, 1, '\t', tabwriter.AlignRight)
			fmt.Fprintln(t, "a\tb\tc\td\t")
			for key, val := range m {
				fmt.Fprintf(t, "%s\t%s\t%s\t%s\n", key, val.Cpu, val.Port, val.State)
			}
			t.Flush()
			stdscr.Printf("%s", b.Bytes())
		}
		stdscr.Refresh()
	}
}

func NewPot(c *DockerCli) *Pot {
	return &Pot{c}
}
