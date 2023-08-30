package main

import (
	"fmt"
	"sync"
	"time"
)

func GenerateInputs(out chan int, wg *sync.WaitGroup, n *OutputStage) {
	defer wg.Done()
	defer close(out)
	for i := 0; i < 100; i++ {
		time.Sleep(time.Nanosecond)

		out <- i
	}
	n.PrevDone = true
}

type OutputStage struct {
	Input  <-chan int
	Buffer []int
	BufEls int
	Output string

	BufLck          sync.Mutex
	BufWriteStalled <-chan bool
	SelfWG          sync.WaitGroup

	BuffererDone bool
	PrevDone     bool
}

func (o *OutputStage) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	o.SelfWG.Add(2)
	go o.Recv()
	go o.Write()
	o.SelfWG.Wait()

	for _, k := range o.Buffer {
		o.Output += fmt.Sprintf("%d, ", k)
	}
}

func (o *OutputStage) Done() bool {
	return (len(o.Input) == 0) && o.PrevDone && o.BuffererDone
}

func (o *OutputStage) Recv() {
	defer o.SelfWG.Done()
	for j := range o.Input {
		for o.BufEls >= 5 {

		}
		o.BufLck.Lock()
		o.Buffer[o.BufEls] = j
		o.BufEls++
		o.BufLck.Unlock()

	}
	o.BuffererDone = true
}
func (o *OutputStage) Write() {
	defer o.SelfWG.Done()
	for !o.Done() {
		if o.BufEls > 2 {
			o.BufLck.Lock()
			o.BufEls--
			j := o.Buffer[o.BufEls]

			o.Output += fmt.Sprintf("%d, ", j)
			o.BufLck.Unlock()
		}
	}
}

func main() {
	i := make(chan int, 100)
	o := OutputStage{Input: i, PrevDone: false, BuffererDone: false, Buffer: make([]int, 5), BufEls: 0}
	w := &sync.WaitGroup{}

	w.Add(2)
	go GenerateInputs(i, w, &o)
	go o.Run(w)

	w.Wait()
	fmt.Println(o.Output)

	fmt.Println("done")

	a := make([]int, 5)
	for p := 0; p < 5; p++ {
		a[p] = p
	}
	var b int
	b, a = a[0], a[1:]
	fmt.Printf("%d\n", b)
	fmt.Printf("%d\n", len(a))
	a[4] = 0
}
