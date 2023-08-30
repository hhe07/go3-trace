package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/muesli/termenv"
)

type Range struct {
	Start uint64
	End   uint64
}

// ! what was tick drift supposed to do
type CLIArgs struct {
	Outfile          string
	TickRange        Range
	InstRange        Range
	Width            uint
	Color            bool
	CycleTime        uint
	Timestamps       bool
	CommitedOnly     bool
	StoreCompletions bool
}

type Stage struct {
	Style     termenv.Style
	Shorthand string
}

var stages = map[string]Stage{
	"fetch":    {termenv.String().Background(termenv.ANSIBrightBlue), "f"},
	"decode":   {termenv.String().Background(termenv.ANSIBrightYellow), "d"},
	"rename":   {termenv.String().Background(termenv.ANSIMagenta), "n"},
	"dispatch": {termenv.String().Background(termenv.ANSIGreen), "p"},
	"issue":    {termenv.String().Background(termenv.ANSIRed), "i"},
	"complete": {termenv.String().Background(termenv.ANSIBrightCyan), "c"},
	"retire":   {termenv.String().Background(termenv.ANSIBlue), "r"},
}

// line fetch and dispatch

type ToOrder struct {
	SN   uint
	Data string
}

func main() {
	fmt.Println("started.")
	c := CLIArgs{
		Outfile:          "wow.txt",
		TickRange:        Range{0, 2000},
		InstRange:        Range{0, 2000},
		Width:            150,
		Color:            true,
		CycleTime:        1000,
		Timestamps:       false,
		CommitedOnly:     true,
		StoreCompletions: false,
	}
	// TODO: check inst ranges / tick ranges
	ords := OrderStage{PrevStageClear: false}
	f, err := os.Open("/home/hhe07/programming/spur/gem5/gem5/stl_o3.txt")
	//f, err := os.Open("/home/hhe07/programming/spur/gem5/gem5/test.txt")
	if err != nil {
		panic(err)
	}

	if c.StoreCompletions {
		stages["store"] = Stage{termenv.String().Background(termenv.ANSIBrightGreen), "s"}
	}
	/* if !c.Color {

	} */
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// ! TODO: consider buffer.Scanner?
	l := LineStage{WorkerCount: 4}
	w := &sync.WaitGroup{}
	l.Init()
	ords.Init(&c, l.LinesOut)

	w.Add(2)
	go l.ProcessLine(scanner, &c, &ords, w)

	go ords.Run(w)
	w.Wait()

}
