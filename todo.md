# further optimisations
- distribute Inst struct creation amongst workerpool
  - this doesn't actually do much, at least not in my testing
- remove string indexing for stage names where possible


after a reset, does a string builder's capacity go to 0? -> no


line length: 10 characters

f: 5
d: 7
n: 7
i: 8
c: 9
r: 11


target:

.....f.dic
.r

procedure:
- get base tick number (0)
- for every event to record:
  - get tick number
  - if tick is less than written, continue
  - write (tick - written) dots
  - write marker


ex:
get tick number 5 -> write (5-0) dots
.....
write marker
.....f

get tick number 7 -> write (7-6) dots
.....f.
write marker
.....f.d


profiler segment

```go
	pf, err := os.Create("prof.txt")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer pf.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(pf); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()
```

lol bnoobjreorder