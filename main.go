package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func customUsage() {
	fmt.Printf("Usage: %s [OPTIONS] results.tsv\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	optMin := flag.Duration("min", time.Duration(-1), "filter rows having at most the given time (3rd column)")
	optMax := flag.Duration("max", time.Duration(-1), "filter rows having at least the given time (3rd column)")
	optGroup := flag.Int("group", -1, "filter by the given group (goroutine) (1st column)")
	optRequest := flag.Int("request", -1, "filter by the given request number (2nd column)")
	optList := flag.Bool("list", false, "list rows, do not display the summary")
	optUnit := flag.String("unit", "ms", "the time unit used for the result, one of: ns, ms (default: ms)")
	flag.Parse()
	flag.Usage = customUsage
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	unit := *optUnit
	if unit != "ns" && unit != "ms" {
		fmt.Fprintln(os.Stderr, "unit must be one of: ns, ms")
		os.Exit(1)
	}
	groupFilter := int64(*optGroup)
	requestFilter := int64(*optRequest)
	limitMin := int64(*optMin)
	limitMax := int64(*optMax)
	canList := *optList
	min := int64(math.MaxInt64)
	var max, total, count int64
	path := flag.Arg(0)
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		panic(fmt.Errorf("error reading file: %s: %w", path, err))
	}
	sc := bufio.NewScanner(f)
	currentLine := 0
	for sc.Scan() {
		currentLine++
		text := sc.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Split(sc.Text(), "\t")
		if len(fields) < 3 {
			panic(fmt.Errorf("invalid line at: %d: %s", currentLine, sc.Text()))
		}
		g := parseField(fields[0], "group", currentLine)
		if groupFilter != -1 && g != groupFilter {
			continue
		}
		req := parseField(fields[1], "request ID", currentLine)
		if requestFilter != -1 && req != requestFilter {
			continue
		}
		ns := parseField(fields[2], "time", currentLine)
		if limitMax >= 0 && ns > limitMax || limitMin >= 0 && ns < limitMin {
			continue
		}
		if canList {
			fmt.Println(sc.Text())
			continue
		}
		if ns > max {
			max = ns
		}
		if ns < min {
			min = ns
		}
		total += ns
		count++
	}
	if !canList && count > 0 {
		label := func(s string) string {
			return fmt.Sprintf("%s (%s)", s, unit)
		}
		avg := float64(total) / float64(count)
		minF := float64(min)
		maxF := float64(max)
		if unit == "ms" {
			avg /= 1_000_000
			minF /= 1_000_000
			maxF /= 1_000_000
		}
		fmt.Printf("%12s\t%12s\t%12s\t%12s\n", "count", label("min"), label("avg"), label("max"))
		fmt.Printf("%12d\t%12.2f\t%12.2f\t%12.2f\n", count, minF, avg, maxF)
	}
}

func parseField(s, name string, line int) int64 {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Errorf("invalid %s at: %d: %s", name, line, s))
	}
	return num
}
