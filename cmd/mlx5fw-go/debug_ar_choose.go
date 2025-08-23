package main

import (
	"os"
	"strconv"
	"strings"
)

// chooseSetStrategy parses MLX5FW_MST_SET_STRATEGY and returns a generator for address_space IDs.
// Supported examples:
//
//	rot:10,2,3   (default rotation)
//	const:3      (always 3)
func chooseSetStrategy() func() uint16 {
	s := os.Getenv("MLX5FW_MST_SET_STRATEGY")
	if s == "" {
		seq := []uint16{10, 2, 3}
		i := 0
		return func() uint16 { v := seq[i%len(seq)]; i++; return v }
	}
	parts := strings.SplitN(s, ":", 2)
	mode := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}
	switch mode {
	case "const":
		n := 3
		if a := strings.TrimSpace(arg); a != "" {
			if parsed, err := strconv.Atoi(a); err == nil {
				n = parsed
			}
		}
		return func() uint16 { return uint16(n) }
	case "rot":
		vals := []uint16{}
		for _, a := range strings.Split(arg, ",") {
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			if parsed, err := strconv.Atoi(a); err == nil {
				vals = append(vals, uint16(parsed))
			}
		}
		if len(vals) == 0 {
			vals = []uint16{10, 2, 3}
		}
		i := 0
		return func() uint16 { v := vals[i%len(vals)]; i++; return v }
	default:
		seq := []uint16{10, 2, 3}
		i := 0
		return func() uint16 { v := seq[i%len(seq)]; i++; return v }
	}
}
