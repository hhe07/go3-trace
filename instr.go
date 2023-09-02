package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

type Instr struct {
	Values   map[StageEnum]*ISort
	Preorder []*ISort
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
			output.WriteString(fmt.Sprintf("	f=%d, r=%d", inst.Values[Fetch].Val, inst.Values[Retire].Val))
		}
	} else {
		output.WriteString("		...")
	}
	output.WriteRune('\n')
}

func BuildInst(a []string, args *CLIArgs) *Instr {

	if len(a) != 7 {
		panic("bad instruction building input")
	}
	res := &Instr{Values: map[StageEnum]*ISort{}}
	var fields = make([]string, 10)
	for _, v := range a {
		//fields := strings.Split(v, ":")
		StaticSplit(v, fields, 10, ":")
		//fmt.Println(fields)

		t, err := strconv.Atoi(fields[2])
		handleConvErr(err)

		if t < int(args.TickRange.Start) || (t > int(args.TickRange.End) && args.TickRange.End != 0) {
			return nil
		}

		res.Values[LookupStage(fields[1])] = &ISort{fields[1], uint(t)}
		if fields[1] == "fetch" {

			s, err := strconv.Atoi(fields[5])
			handleConvErr(err)
			if s < int(args.InstRange.Start) || (s > int(args.InstRange.End) && args.InstRange.End != 0) {
				return nil
			}

			pc, err := strconv.ParseInt(fields[3], 0, 64)
			handleConvErr(err)

			upc, err := strconv.Atoi(fields[4])
			handleConvErr(err)

			res.PC = uint(pc)
			res.UPC = uint(upc)
			res.Disasm = fields[6]
			res.SN = uint(s)
		}
	}
	return res
}

func StaticSplit(in string, out []string, times int, sep string) {
	var b string
	var found bool
	var i int
	for i = 0; i < times; i++ {
		b, in, found = strings.Cut(in, sep)
		if !found {
			out[i] = b
			break
		}
		out[i] = b
	}
}

func LookupStage(name string) StageEnum {
	switch name {
	case "fetch":
		return Fetch
	case "decode":
		return Decode
	case "rename":
		return Rename
	case "dispatch":
		return Dispatch
	case "issue":
		return Issue
	case "complete":
		return Complete
	case "retire":
		return Retire
	case "store":
		return Store
	default:
		return 99
	}
}
