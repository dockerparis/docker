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

const NB_COLUMNS = 8
const COLOR_CONTAINER = 2

type Status int			// What we are doing

const (
	STATUS_POT = iota	// Currently displaying containers
	STATUS_POT_ALL		// Currently displaying containers and processes
	STATUS_HELP		// Currently displaying help
	STATUS_CONFIRM		// Currently waiting for confirmation
)

type Pot struct {
	c *DockerCli		// Used to talk to the daemon
	status Status		// Current status
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

func PrettyColumn(in string, expected_len int) string {
	prefix := " "
	suffix := " |"

	i := len(in) + len(prefix) + len(suffix)
	if (i < expected_len) {
		return prefix + in + strings.Repeat(" ", expected_len - i) + suffix
	}
	if (i > expected_len) {
		j := expected_len - len(prefix) - len(suffix)
		return prefix + in[0:j] + suffix
	}

	return prefix + in + suffix
}

func (c *ContainerLine) Format(column_width int) string {
	return PrettyColumn(c.Name, column_width) +
		PrettyColumn(c.Image, column_width) +
		PrettyColumn(c.Id, column_width) +
		PrettyColumn(c.Command, column_width) +
		PrettyColumn(c.Uptime, column_width) +
		PrettyColumn(c.Status, column_width) +
		PrettyColumn(c.CPU, column_width) +
		PrettyColumn(c.RAM, column_width)
}

// ProcessLine contains information about a process
type ProcessLine ContainerLine

func (c *ProcessLine) Format(column_width int) string {
	return PrettyColumn("", column_width) +
		PrettyColumn("", column_width) +
		PrettyColumn(c.Id, column_width) +
		PrettyColumn(c.Command, column_width) +
		PrettyColumn(c.Uptime, column_width) +
		PrettyColumn(c.Status, column_width) +
		PrettyColumn(c.CPU, column_width) +
		PrettyColumn(c.RAM, column_width)
}

type Container struct {
	container ContainerLine // information about the container
	processes []ProcessLine // information about the processes
}

type PrintedLine struct {
	line string // the line
	isContainer bool // is this line a container?
	isProcess bool // is this line a process?
}

var (
	active         = 0
	scroll         = 0
)

// Returns the running processes for the current Container
func (pot *Pot) GetProcesses(cid string) []ProcessLine {
	res := make([]ProcessLine, 0, 0)
	val := url.Values{}
	val.Set("ps_args", "-o pid,etime,%cpu,%mem,cmd")
	stream, _, err := pot.c.call("GET", "/containers/"+cid+"/top?"+val.Encode(), nil, false)
	if err != nil {
		return res
	}
	var procs engine.Env
	if err := procs.Decode(stream); err != nil {
		return res
	}
	processes := [][]string{}
	if err := procs.GetJson("Processes", &processes); err != nil {
		return res
	}
	for _, proc := range processes {
		var p ProcessLine

		p.Id = proc[0]
		p.Uptime = proc[1]
		p.CPU = proc[2]
		p.RAM = proc[3]
		p.Command = proc[4]

		res = append(res, p)		
	}
	
	return res
}

// Returns the list of running containers as well as internal processes
func (pot *Pot) Snapshot() []Container {
	res := make([]Container, 0, 16)

	v := url.Values{}
	v.Set("all", "1")
	body, _, err := readBody(pot.c.call("GET", "/containers/json?"+v.Encode(), nil, false))
	if err != nil {
		return res
	}
	outs := engine.NewTable("Created", 0)
	if _, err = outs.ReadListFrom(body); err != nil {
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

		c.processes = pot.GetProcesses(c.container.Id)

		total_cpu := 0.0
		total_ram := 0.0
		for _, p := range(c.processes) {
			cpu, err := strconv.ParseFloat(p.CPU, 32)
			if err == nil {
				total_cpu = total_cpu + cpu
			}
			ram, err := strconv.ParseFloat(p.RAM, 32)
			if err == nil {
				total_ram = total_ram + ram
			}
		}
		c.container.CPU = fmt.Sprintf("%.1f", total_cpu)
		c.container.RAM = fmt.Sprintf("%.1f", total_ram)

		res = append(res, c)
	}

	return res
}

func (pot *Pot) PrintActive(win *gnc.Window, l PrintedLine, lc int, i int) {
	if i < scroll || i >= scroll+lc {
		return
	}
	if active == i {
		win.AttrOn(gnc.A_REVERSE)
		win.Println(l.line)
		win.AttrOff(gnc.A_REVERSE)
	} else {
		if l.isContainer {
			win.ColorOn(COLOR_CONTAINER)
			win.Println(l.line)
			win.ColorOff(COLOR_CONTAINER)
		} else {
			win.Println(l.line)
		}
	}
}

func (pot *Pot) Update(win *gnc.Window, lc int, wc int, cnts []Container) {
	ss := make([]PrintedLine, 0, 42)
	for _, cnt := range cnts {
		ss = append(ss, PrintedLine{cnt.container.Format(wc), true, false})
		for _, proc := range cnt.processes {
			ss = append(ss, PrintedLine{proc.Format(wc), false, true})
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
		pot.PrintActive(win, s, lc, i)
	}
}

func (pot *Pot) Run() {
	win, err := gnc.Init()

	gnc.StartColor()
	gnc.InitPair(COLOR_CONTAINER, gnc.C_BLACK, gnc.C_CYAN)
	
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

	recompute := false
	snapshot := pot.Snapshot()

	for {
		my, mx := win.MaxYX()
	
		lc := my - 2 // size max of y - header (1)
		wc := (mx - 1) / NB_COLUMNS
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
			recompute = true
		case <-s:
			recompute = true
			gnc.End()
			win.Refresh()
		}
		win.Erase()
		o, _ := exec.Command("uptime").Output()
		win.Printf("%s", o)

		win.AttrOn(gnc.A_REVERSE)

		win.Println(PrettyColumn("Name", wc) +
			PrettyColumn("Image", wc) +
			PrettyColumn("Id", wc) +
			PrettyColumn("Command", wc) +
			PrettyColumn("Uptime", wc) +
			PrettyColumn("Status", wc) +
			PrettyColumn("%CPU", wc) +
			PrettyColumn("%RAM", wc))
		
		win.AttrOff(gnc.A_REVERSE)
		
		if mx < 40 || my < 5 {
			continue
		}
		if recompute {
			snapshot = pot.Snapshot()
			recompute = false
		}
		pot.Update(win, lc, wc, snapshot)
		win.Refresh()
	}
}

func NewPot(c *DockerCli) *Pot {
	return &Pot{c, STATUS_POT}
}
