package pot

import (
	"fmt"
	"time"
)

type Pot struct {
}

func (pot *Pot) Run() {
	for {
		fmt.Printf("loop\n")
		time.Sleep(1 * 1e9)
	}
}

func NewPot() *Pot {
	return &Pot{}
}
