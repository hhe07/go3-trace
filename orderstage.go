package main

import (
	"bufio"
	"os"
	"sort"
	"sync"
)

type OrderStage struct {
	PrintQueue     []*ToOrder // queue of instructions to print out
	Outfile        *os.File
	OutHandle      bufio.Writer
	In             <-chan *ToOrder
	PrevStageClear bool
}

func (o *OrderStage) sortfunc(a, b int) bool {
	return o.PrintQueue[a].SN < o.PrintQueue[b].SN
}

func (o *OrderStage) Init(a *CLIArgs, i <-chan *ToOrder) {
	var err error
	o.In = i
	o.Outfile, err = os.Create(a.Outfile)
	if err != nil {
		panic("error on creating output file")
	}
	o.OutHandle = *bufio.NewWriter(o.Outfile)
}

func (o *OrderStage) Done() bool {
	return (len(o.In) == 0) && o.PrevStageClear
}

func (o *OrderStage) Run(w *sync.WaitGroup) {
	defer w.Done()
	for i := range o.In {
		if len(o.PrintQueue) == 2000 {
			sort.Slice(o.PrintQueue, o.sortfunc)
			for j := 0; j < 1500; j++ {
				o.OutHandle.WriteString(o.PrintQueue[j].Data)
			}
			o.PrintQueue = o.PrintQueue[1500:]
		}
		o.PrintQueue = append(o.PrintQueue, i)
	}
	o.Clean()

}

func (o *OrderStage) Clean() {
	// sort array
	sort.Slice(o.PrintQueue, o.sortfunc)
	// write out remaining
	for _, v := range o.PrintQueue {
		o.OutHandle.WriteString(v.Data)
	}
	o.OutHandle.Flush()
	o.Outfile.Close()
}
