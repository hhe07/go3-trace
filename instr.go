package main

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

type Instr struct {
	Values   map[string]ISort
	Preorder []ISort
	PC       uint
	UPC      uint
	Disasm   string
	SN       uint
}

// ticks are the first time of an event

type ISort struct {
	Name string
	Val  uint
}

func (i *Instr) LastEvent(args *CLIArgs) uint {
	i.Preorder = nil
	i.Preorder = maps.Values(i.Values)
	lf := func(a, b int) bool { return i.Preorder[a].Val < i.Preorder[b].Val }
	sort.Slice(i.Preorder, lf)
	return i.Preorder[len(i.Preorder)-1].Val
}

func (inst Instr) handleLineEnd(output *strings.Builder, ln uint, args *CLIArgs, tickRepr uint) {
	output.WriteString(fmt.Sprintf("]-(%15d)", tickRepr))
	if ln == 1 {
		output.WriteString(fmt.Sprintf("%10x.%d %25s [%10d]", inst.PC, inst.UPC, inst.Disasm, inst.SN))
		if args.Timestamps {
			output.WriteString(fmt.Sprintf("	f=%d, r=%d", inst.Values["fetch"].Val, inst.Values["retire"].Val))
		}
	} else {
		output.WriteString("		...")
	}
	output.WriteRune('\n')
}
