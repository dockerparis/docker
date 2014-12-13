package pot

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"
	"time"

	"code.google.com/p/goncurses"
)

type Pot struct{}

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

func NewPot() *Pot {
	return &Pot{}
}
