package main

import (
	"bufio"
	"strconv"
	"strings"
	"sync"

	"github.com/muesli/termenv"
)

type LineStage struct {
	WDist       chan *Instr // channel for handing off instructions to workers
	LinesOut    chan *ToOrder
	WorkerCount int
	Workers     []LineWorker
}

type LineWorker struct {
	InstIn       <-chan *Instr
	LinesOut     chan<- *ToOrder
	StringConstr strings.Builder
	WG           *sync.WaitGroup
}

func (l *LineWorker) Init(input <-chan *Instr, output chan<- *ToOrder, wg *sync.WaitGroup) {
	l.InstIn = input
	l.LinesOut = output
	l.WG = wg
}

func (l *LineWorker) Run(args *CLIArgs) {
	defer l.WG.Done()
	var dot byte = '.'
	var currentStyle termenv.Style
	var timeWidth = uint(args.Width * args.CycleTime) // number of cycles which can be represented on a given line

	for inst := range l.InstIn {
		currentStyle = termenv.Style{}
		l.StringConstr.Reset()
		l.StringConstr.Grow(int(args.Width))

		fetch := inst.Values["fetch"].Val

		baseTick := (fetch / timeWidth) * timeWidth // set basetick to the lowest multiple of timeWidth
		lastEvent := inst.LastEvent(args)           // lastEvent is in raw ticks
		numLines := (lastEvent-fetch)/timeWidth + 1
		dot = '.'
		if inst.Values["retire"].Val == 0 {
			dot = '='
		}

		l.StringConstr.WriteRune('[')
		var written uint = 0
		var currentLine uint = 1

		for _, event := range inst.Preorder {
			var subconstr strings.Builder
			// adjust event time using baseTick
			if event.Val == 0 {
				continue
			}
			event.Val -= baseTick
			for written < (event.Val / uint(args.CycleTime)) {
				subconstr.WriteRune(rune(dot))
				written++
				// ! TODO: handle compact case

				// if written up to line limit,
				if written >= args.Width*currentLine {
					// immediately push and reset subconstr
					l.StringConstr.WriteString(currentStyle.Styled(subconstr.String()))
					subconstr.Reset()

					// handle line end
					inst.handleLineEnd(&l.StringConstr, currentLine, args, baseTick+uint(currentLine)*timeWidth)
					l.StringConstr.WriteRune('[')

					currentLine++
				}
			}
			// push leading
			l.StringConstr.WriteString(currentStyle.Styled(subconstr.String()))
			written++
			subconstr.Reset()

			// get a new style, write marker
			currentStage := stages[event.Name]
			currentStyle = currentStage.Style
			l.StringConstr.WriteString(currentStyle.Styled(currentStage.Shorthand))
		}
		// write remainder of last line, if necessary
		for written < args.Width*numLines {
			l.StringConstr.WriteRune(rune(dot))
			written++
		}
		inst.handleLineEnd(&l.StringConstr, currentLine, args, baseTick+uint(currentLine)*timeWidth)

		l.LinesOut <- &ToOrder{inst.SN, l.StringConstr.String()}
	}
}

func (l *LineStage) Init() {
	l.WDist = make(chan *Instr, 100)
	l.LinesOut = make(chan *ToOrder, 400)
}

func (l *LineStage) ProcessLine(scanner *bufio.Scanner, args *CLIArgs, o *OrderStage, w *sync.WaitGroup) {
	wg := &sync.WaitGroup{}
	l.Workers = make([]LineWorker, l.WorkerCount)
	for w := 0; w < l.WorkerCount; w++ {
		wg.Add(1)
		l.Workers[w].Init(l.WDist, l.LinesOut, wg)
		go l.Workers[w].Run(args)
	}

	var i *Instr
	var fields []string

	for scanner.Scan() {
		cl := scanner.Text()
		if len(cl) < 10 {
			continue
		}
		fields = strings.SplitN(cl, ":", 10)
		if fields[0] != "O3PipeView" || len(fields) == 1 {
			continue // immediately discard
		}

		t, err := strconv.Atoi(fields[2])
		handleConvErr(err)

		if t < int(args.TickRange.Start) {
			continue
		}
		if fields[1] == "fetch" {

			s, err := strconv.Atoi(fields[5])
			handleConvErr(err)

			if fields[1] == "fetch" && s < int(args.InstRange.Start) {
				continue
			}
			pc, err := strconv.ParseInt(fields[3], 0, 64)
			handleConvErr(err)

			upc, err := strconv.Atoi(fields[4])
			handleConvErr(err)

			i = &Instr{
				Values: map[string]ISort{"fetch": {"fetch", uint(t)}},
				PC:     uint(pc),
				UPC:    uint(upc),
				Disasm: fields[6],
				SN:     uint(s),
			}
		} else {
			i.Values[fields[1]] = ISort{fields[1], uint(t)}
			if fields[1] == "retire" {
				l.WDist <- i
			}
		}
	}
	close(l.WDist)
	wg.Wait()
	o.PrevStageClear = true
	close(l.LinesOut)
	w.Done()

}

func handleConvErr(e error) {
	if e != nil {
		panic("bad conversion")
	}
}
