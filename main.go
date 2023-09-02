package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"
)

type Range struct {
	Start uint64
	End   uint64
}

func (r *Range) Valid() bool {
	return r.Start <= r.End
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
	CommittedOnly    bool
	StoreCompletions bool
}

type Stage struct {
	ColourIntro string
	Shorthand   string
}

type StageEnum uint8

const (
	Fetch StageEnum = iota
	Decode
	Rename
	Dispatch
	Issue
	Complete
	Retire
	Store
)

var stages = map[string]*Stage{ // ! please make sure to keep the colour strings are kept at a constant length
	"fetch":    {ANSIIntro + "[48;5;001m", "f"},
	"decode":   {ANSIIntro + "[48;5;002m", "d"},
	"rename":   {ANSIIntro + "[48;5;003m", "n"},
	"dispatch": {ANSIIntro + "[48;5;004m", "p"},
	"issue":    {ANSIIntro + "[48;5;005m", "i"},
	"complete": {ANSIIntro + "[48;5;006m", "c"},
	"retire":   {ANSIIntro + "[48;5;009m", "r"},
}

const ANSIIntro = string('\x1b')
const ANSITerminator = ANSIIntro + "[0m"

// line fetch and dispatch

type ToOrder struct {
	SN   uint
	Data string
}

func main() {
	c := CLIArgs{
		TickRange: Range{0, 0},
		InstRange: Range{0, 0},
	}

	var infilename string

	// TODO: implement committedOnly functionality

	flag.StringVar(&c.Outfile, "outfile", "o3-pipeview.out", "filename of output")

	flag.UintVar(&c.CycleTime, "cycleTime", 1000, "ticks per cycle")
	flag.UintVar(&c.Width, "width", 150, "width of tick reprs per line")

	flag.Uint64Var(&c.TickRange.Start, "tickMin", 0, "lowest tick to print (inclusive)")
	flag.Uint64Var(&c.TickRange.End, "tickMax", 0, "highest tick to print (inclusive). 0 means no restriction.")
	flag.Uint64Var(&c.InstRange.Start, "instMin", 0, "lowest SN to print (inclusive)")
	flag.Uint64Var(&c.InstRange.End, "instMax", 0, "highest SN to print (inclusive). 0 means no restriction.")

	flag.BoolVar(&c.Color, "color", true, "enable colour")
	flag.BoolVar(&c.Timestamps, "timestamp", false, "print timestamps")
	flag.BoolVar(&c.CommittedOnly, "committed", false, "print committed only")
	flag.BoolVar(&c.StoreCompletions, "storeCompletions", false, "store completions")

	flag.Parse()

	if !c.TickRange.Valid() {
		fmt.Println("bad tick range")
		return
	}

	if !c.InstRange.Valid() {
		fmt.Println("bad SN range")
		return
	}

	infilename = flag.Arg(0)
	ords := OrderStage{PrevStageClear: false}
	f, err := os.Open(infilename)
	if err != nil {
		panic(err)
	}

	if c.StoreCompletions {
		stages["store"] = &Stage{ANSIIntro + "[0:5:10", "s"}
	}
	if !c.Color {
		for _, v := range stages {
			v.ColourIntro = ""
		}
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	l := LineStage{WorkerCount: 8}
	w := &sync.WaitGroup{}
	l.Init()
	ords.Init(&c, l.LinesOut)

	fmt.Println("started.")
	w.Add(2)
	go l.ProcessLine(scanner, &c, &ords, w)

	go ords.Run(w)
	w.Wait()
	fmt.Println("finished.")
}
