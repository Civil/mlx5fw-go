package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type op struct {
	kind string // set|get|readbuf|writebuf|dword
	size int
}

func parseStrace(path string) ([]op, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ops := []op{}
	// Match _IOC(_IOC_WRITE, 0xd2, 5, 0x10c) and _IOC(_IOC_READ, 0xd2, 4, 0x10c)
	re := regexp.MustCompile(`_IOC\(_IOC_(READ|WRITE),\s*0x[dD]2,\s*([0-9x]+),\s*0x([0-9a-fA-F]+)\)`)
	// Return value exposes bytes transferred for block ops; not strictly needed
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		dir := m[1]
		nr := strings.ToLower(m[2])
		// sizeHex := m[3]
		// Block buffer IOCTLs use size 0x10c, address_space header is 12B, data up to 256
		// We infer data length from return code when present; otherwise approximate by common sizes
		// Here, try to parse the return value at end `= N` if present
		sz := 0
		if i := strings.LastIndex(line, ") = "); i >= 0 {
			tail := strings.TrimSpace(line[i+4:])
			if n, err := strconv.Atoi(strings.TrimPrefix(tail, "+")); err == nil {
				sz = n
			}
		}
		kind := ""
		nrv := 0
		if strings.HasPrefix(nr, "0x") {
			if v, err := strconv.ParseInt(nr, 0, 32); err == nil {
				nrv = int(v)
			}
		} else {
			if v, err := strconv.Atoi(nr); err == nil {
				nrv = v
			}
		}
		if nrv == 5 && dir == "WRITE" {
			kind = "writebuf"
		}
		if nrv == 4 && dir == "READ" {
			kind = "readbuf"
		}
		if nrv == 8 && dir == "WRITE" {
			kind = "set"
		}
		if nrv == 7 && dir == "READ" {
			kind = "get"
		}
		if kind == "" {
			continue
		}
		ops = append(ops, op{kind: kind, size: sz})
	}
	return ops, nil
}

func parseOurLog(path string) ([]op, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ops := []op{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "mst.ioctl.write4_buffer") {
			if sz := extractIntField(line, "size"); sz > 0 {
				ops = append(ops, op{kind: "writebuf", size: sz})
			}
		} else if strings.Contains(line, "mst.ioctl.read4_buffer") {
			if sz := extractIntField(line, "size"); sz > 0 {
				ops = append(ops, op{kind: "readbuf", size: sz})
			}
		} else if strings.Contains(line, "mst.set_space.begin") {
			ops = append(ops, op{kind: "set", size: 0})
		} else if strings.Contains(line, "mst.get_space") {
			ops = append(ops, op{kind: "get", size: 0})
		}
	}
	return ops, nil
}

func extractIntField(line, key string) int {
	// Try key=value
	re1 := regexp.MustCompile(key + "=([0-9]+)")
	if m := re1.FindStringSubmatch(line); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}
	// Try JSON-ish: "key": <num>
	// Note: Go string literals do not recognize \s, so we must escape backslashes for regex
	re2 := regexp.MustCompile("\"" + key + "\"\\s*:\\s*([0-9]+)")
	if m := re2.FindStringSubmatch(line); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return n
		}
	}
	return 0
}

func summarize(ops []op) map[string]map[int]int {
	res := map[string]map[int]int{}
	for _, o := range ops {
		if _, ok := res[o.kind]; !ok {
			res[o.kind] = map[int]int{}
		}
		res[o.kind][o.size]++
	}
	return res
}

func createDebugMSTDiffCommand() *cobra.Command {
	var tracePath, oursPath string
	var max int
	cmd := &cobra.Command{
		Use:   "mst-diff",
		Short: "Compare MST IOCTL cadence with a flint strace",
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := parseStrace(tracePath)
			if err != nil {
				return fmt.Errorf("parse trace: %w", err)
			}
			ours, err := parseOurLog(oursPath)
			if err != nil {
				return fmt.Errorf("parse ours: %w", err)
			}
			fmt.Printf("Ref ops=%d, Ours ops=%d\n", len(ref), len(ours))
			rsum := summarize(ref)
			osum := summarize(ours)
			fmt.Println("Ref histogram (readbuf/writebuf sizes):")
			for k, m := range rsum {
				if k == "readbuf" || k == "writebuf" {
					for sz, c := range m {
						fmt.Printf("  %s %3d => %d\n", k, sz, c)
					}
				}
			}
			fmt.Println("Ours histogram (readbuf/writebuf sizes):")
			for k, m := range osum {
				if k == "readbuf" || k == "writebuf" {
					for sz, c := range m {
						fmt.Printf("  %s %3d => %d\n", k, sz, c)
					}
				}
			}
			// First 50 ops side-by-side
			lim := max
			if lim <= 0 || lim > len(ref) {
				lim = len(ref)
			}
			if lim > len(ours) {
				lim = len(ours)
			}
			fmt.Printf("First %d ops side-by-side (kind,size):\n", lim)
			for i := 0; i < lim; i++ {
				ro, oo := ref[i], ours[i]
				mark := ""
				if ro.kind != oo.kind || (ro.size != 0 && oo.size != 0 && ro.size != oo.size) {
					mark = " <== diff"
				}
				fmt.Printf("%3d: ref(%s,%d) ours(%s,%d)%s\n", i, ro.kind, ro.size, oo.kind, oo.size, mark)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tracePath, "trace", "", "Path to flint_oem strace file")
	cmd.Flags().StringVar(&oursPath, "ours", "", "Path to our CLI verbose log")
	cmd.Flags().IntVar(&max, "max", 80, "Max ops to print side-by-side")
	cmd.MarkFlagRequired("trace")
	cmd.MarkFlagRequired("ours")
	return cmd
}
