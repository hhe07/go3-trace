package main

import (
	"bufio"
	"bytes"
	"strings"
	"sync"
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
	var colour string

	var dots []byte // when true, skip to '=' dots
	var dotBuffer []byte = append(bytes.Repeat([]byte{'.'}, 150), bytes.Repeat([]byte{'='}, 150)...)
	var timeWidth = args.Width * args.CycleTime // number of cycles which can be represented on a given line
	var subconstr strings.Builder

	for inst := range l.InstIn {
		var stw string

		if args.CommittedOnly && inst.Values[Retire].Val == 0 {
			continue
		}
		// reset variables
		colour = ""

		l.StringConstr.Reset()

		fetch := inst.Values[Fetch].Val

		baseTick := (fetch / timeWidth) * timeWidth // set basetick to the lowest multiple of timeWidth
		lastEvent := inst.LastEvent(args)           // lastEvent is in raw ticks
		numLines := (lastEvent-fetch)/timeWidth + 1 // number of lines required

		dots = dotBuffer[:150]
		if inst.Values[Fetch].Val == 0 {
			dots = dotBuffer[150:]
		}

		l.StringConstr.WriteRune('[')
		var written uint = 0
		var currentLine uint = 1

		l.StringConstr.Grow(int(args.Width*numLines) + 250) // some amount of padding for formatting characters

		for _, event := range inst.Preorder {
			subconstr.Reset()
			subconstr.Grow(int(args.Width))

			if event.Val == 0 {
				continue
			}
			// adjust event time using baseTick
			event.Val -= baseTick

			eTime := (event.Val / args.CycleTime)
			if eTime < written {
				continue
			}
			dotsToWrite := eTime - written
			lineLim := args.Width * currentLine
			for dotsToWrite > 0 {
				// if going to overflow line,
				if (written + dotsToWrite) >= lineLim {

					t := lineLim - written

					subconstr.Write(dots[:t])
					dotsToWrite -= t
					written += t

					writeWithColour(colour, subconstr.String(), &l.StringConstr)
					subconstr.Reset()

					// handle line end
					inst.handleLineEnd(&l.StringConstr, currentLine, args, baseTick+currentLine*timeWidth)
					l.StringConstr.WriteRune('[')

					currentLine++
					lineLim += args.Width
				} else {
					subconstr.Write(dots[:dotsToWrite])
					written += dotsToWrite
					dotsToWrite = 0
				}
			}

			writeWithColour(colour, subconstr.String(), &l.StringConstr)
			subconstr.Reset()

			// get a new style, write marker
			currentStage := stages[event.Name]
			colour = currentStage.ColourIntro
			// push leading
			writeWithColour(colour, currentStage.Shorthand, &l.StringConstr)
			written++

			// if the above pushed us to line limit, handle
			if written == lineLim {
				inst.handleLineEnd(&l.StringConstr, currentLine, args, baseTick+currentLine*timeWidth)
				l.StringConstr.WriteRune('[')
				currentLine++
				lineLim += args.Width
			}
		}

		// write remainder of last line, if necessary
		if written <= args.Width*numLines {
			e := args.Width*numLines - written
			l.StringConstr.Write(dots[:e])
			written += e
		}

		if currentLine > numLines && numLines == 1 {
			s := strings.Split(l.StringConstr.String(), "\n")
			l := len(s[1]) - strings.Count(s[1], ANSITerminator)*4 - strings.Count(s[1], ANSIIntro+"[48")*11
			stw = s[1] + s[0][l:] + "\n"
			currentLine--
		} else {
			inst.handleLineEnd(&l.StringConstr, currentLine, args, baseTick+currentLine*timeWidth)
			stw = l.StringConstr.String()
		}

		l.LinesOut <- &ToOrder{inst.SN, stw}
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

	s := make([]string, 7)
	ct := 0
	for scanner.Scan() {
		cl := scanner.Text()
		if len(cl) < 10 || cl[:10] != "O3PipeView" {
			continue
		}
		s[ct] = cl
		ct++
		if ct == 7 {
			inst := BuildInst(s, args)
			if inst == nil {
				ct = 0
				continue
			}
			l.WDist <- inst
			s = make([]string, 7)
			ct = 0
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

func writeWithColour(colour, b string, mb *strings.Builder) {
	mb.WriteString(colour)
	mb.WriteString(b)
	mb.WriteString(ANSITerminator)
}
